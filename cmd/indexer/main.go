package main

import (
	"context"
	"flow-indexer/internal/adapter"
	"flow-indexer/internal/domain/account"
	flowEvent "flow-indexer/internal/domain/event"
	"flow-indexer/internal/domain/inscription"
	"flow-indexer/internal/service"
	"flow-indexer/pkg/log"
	"fmt"
	"sync"
	"time"

	flowUtils "flow-indexer/pkg/flow"

	"github.com/onflow/flow-go-sdk/client"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"

	gormpkg "flow-indexer/pkg/gorm"
)

func main() {
	// init logger
	syncFun, err := log.Init(log.Config{
		Name:   "indexer.log",
		Level:  zapcore.ErrorLevel,
		Stdout: true,
		// File:   "log/indexer/indexer.log",
		File: "",
	})
	if err != nil {
		panic(err)
	}
	defer syncFun()
	logger := zap.L()

	// prepare context
	// ctx := app.GraceCtx(context.Background())

	// init db
	time.Sleep(1 * time.Second)
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		"abc", "abc", "db", "5432", "postgres")
	logger.Info("dsn", zap.String("dsn", dsn))

	db, err := gormpkg.NewGormPostgresConn(
		gormpkg.Config{
			DSN:             dsn,
			MaxIdleConns:    2,
			MaxOpenConns:    2,
			ConnMaxLifetime: 10 * time.Minute,
			SingularTable:   true,
		},
	)
	if err != nil {
		logger.Error("connect to database error", zap.Error(err))
		return
	}

	// create extension
	db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"")

	// migrate db
	err = db.AutoMigrate(
		&account.Account{},
		&inscription.Balance{},
		&flowEvent.FlowEvent{},
	)
	if err != nil {
		logger.Error("migrate db error", zap.Error(err))
		return
	}

	// prepare service
	accountRepo := adapter.NewAccountRepo(db)
	inscriptionRepo := adapter.NewInscriptionRepo(db)
	eventRepo := adapter.NewEventRepo(db)

	svc := service.NewService(
		accountRepo,
		inscriptionRepo,
		eventRepo,
	)

	// init flow client
	maxMsgSize := 50 * 1024 * 1024 // 50MB
	flowClient, err := client.New(
		"access.mainnet.nodes.onflow.org:9000",
		grpc.WithInsecure(),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(maxMsgSize)),
	)

	if err != nil {
		panic(err)
	}

	// scan
	thread := 20
	maxBlockQuery := uint64(250) - 1
	startBlock := uint64(68277132) // freeflow deployment block
	// latestBlock, err := flowClient.GetLatestBlock(context.Background(), true)
	// if err != nil {
	// 	panic(err)
	// }
	// endBlock := latestBlock.Height
	endBlock := uint64(69106534)

	fmt.Println(startBlock, endBlock)
	blockRanges := getBlockRanges(startBlock, endBlock, uint64(thread))
	var wg sync.WaitGroup
	for _, blockRange := range blockRanges {
		wg.Add(1)
		fmt.Println(blockRange.startBlock, blockRange.endBlock)
		scanRangeEvents(blockRange.startBlock, blockRange.endBlock, maxBlockQuery, flowClient, logger, svc, &wg)
	}

	wg.Wait()
}

type blockRange struct {
	startBlock uint64
	endBlock   uint64
}

func getBlockRanges(startBlock, endBlock, thread uint64) []blockRange {
	totalBlocks := endBlock - startBlock + 1
	groupSize := totalBlocks / thread
	if groupSize == 0 {
		groupSize = 1
	}

	var blockRanges []blockRange
	for i := startBlock; i <= endBlock; i += groupSize {
		end := i + groupSize - 1
		if end > endBlock {
			end = endBlock
		}

		blockRanges = append(blockRanges, blockRange{
			startBlock: i,
			endBlock:   end,
		})

		if end == endBlock {
			break
		}
	}

	return blockRanges
}

func scanRangeEvents(
	startBlock, endBlock, maxBlockQuery uint64,
	flowClient *client.Client,
	logger *zap.Logger,
	svc service.Service,
	wg *sync.WaitGroup,
) {
	go func() {
		for i := startBlock; i <= endBlock; i += maxBlockQuery {
			end := i + maxBlockQuery - 1
			if end > endBlock {
				end = endBlock
			}

			scanBatchEvents(i, end, flowClient, logger, svc)
		}
		defer wg.Done()
	}()
}

func scanBatchEvents(
	startBlock, endBlock uint64, flowClient *client.Client, logger *zap.Logger, svc service.Service,
) {
	freeflowDepositEventType := "A.88dd257fcf26d3cc.Inscription.Deposit"
	bes, err := flowClient.GetEventsForHeightRange(context.Background(),
		client.EventRangeQuery{
			Type:        freeflowDepositEventType,
			StartHeight: startBlock,
			EndHeight:   endBlock,
		})
	if err != nil {
		logger.Error("GetEventsForHeightRange", zap.Error(fmt.Errorf("range %x - %x", startBlock, endBlock)))
		return
	}

	for _, be := range bes {
		logger.Info("BlockEvent", zap.Uint64("BlockHeight", be.Height))
		for _, e := range be.Events {
			if e.Type != freeflowDepositEventType {
				continue
			}
			logger.Info("Event", zap.String("Type", e.Type))
			logger.Info("Event", zap.String("TransactionID", e.TransactionID.String()))
			logger.Info("Event", zap.String("TransactionIndex", fmt.Sprintf("%d", e.TransactionIndex)))
			logger.Info("Event", zap.String("EventIndex", fmt.Sprintf("%d", e.EventIndex)))

			flowEvent := flowUtils.FreeflowDeposit(e)
			logger.Info("Event", zap.Uint64("ID", flowEvent.ID()))
			logger.Info("Event", zap.String("Address", fmt.Sprintf("%x", flowEvent.Address())))

			err := svc.CreateFlowEvent(context.Background(), fmt.Sprintf("%x", flowEvent.Address()), e.Type, be.Height)
			if err != nil {
				logger.Error("CreateFlowEvent", zap.Error(err))
				return
			}

			err = svc.UpdateBalance(context.Background(), "freeflow", fmt.Sprintf("%x", flowEvent.Address()), true)
			if err != nil {
				logger.Error("UpdateBalance", zap.Error(
					fmt.Errorf(
						"address: %x, height: %x, range %x - %x", flowEvent.Address(), be.Height, startBlock, endBlock,
					),
				))
				return
			}
		}
	}
}

func getBlockTxs(
	blockNum uint64,
	flowClient *client.Client,
	logger *zap.Logger,
	svc service.Service,
) {
	ctx := context.Background()
	logger.Info("Block", zap.Uint64("BlockHeight", blockNum))
	block, err := flowClient.GetBlockByHeight(ctx, blockNum)
	if err != nil {
		panic(err)
	}

	for _, c := range block.CollectionGuarantees {
		col, err := flowClient.GetCollection(ctx, c.CollectionID)
		if err != nil {
			logger.Error("GetCollection", zap.Error(err))
		}

		for _, tID := range col.TransactionIDs {
			// logger.Info("TransactionID", zap.String("ID", tID.String()))
			tx, err := flowClient.GetTransactionResult(ctx, tID)
			if err != nil {
				logger.Error("GetTransaction", zap.Error(err))
			}
			// logger.Info("Transaction", zap.String("ID", tx.TransactionID.String()))
			// logger.Info("Transaction", zap.String("Status", tx.Status.String()))

			for _, e := range tx.Events {
				// logger.Info("Event", zap.String("Type", e.Type))
				if e.Type != "A.88dd257fcf26d3cc.Inscription.Deposit" {
					continue
				}
				// logger.Info("Event", zap.String("TransactionID", e.TransactionID.String()))
				// logger.Info("Event", zap.String("TransactionIndex", fmt.Sprintf("%d", e.TransactionIndex)))
				// logger.Info("Event", zap.String("EventIndex", fmt.Sprintf("%d", e.EventIndex)))

				flowEvent := flowUtils.FreeflowDeposit(e)
				// logger.Info("Event", zap.Uint64("ID", flowEvent.ID()))
				// logger.Info("Event", zap.String("Address", fmt.Sprintf("%x", flowEvent.Address())))

				err := svc.UpdateBalance(ctx, "freeflow", fmt.Sprintf("%x", flowEvent.Address()), true)
				if err != nil {
					logger.Error("UpdateBalance", zap.Error(err))
					return
				}
			}
		}
	}

}
