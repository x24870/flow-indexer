package main

import (
	"context"
	"flow-indexer/internal/adapter"
	"flow-indexer/internal/domain/account"
	"flow-indexer/internal/domain/inscription"
	"flow-indexer/internal/service"
	"flow-indexer/pkg/log"
	"fmt"
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
	sync, err := log.Init(log.Config{
		Name:   "indexer.log",
		Level:  zapcore.DebugLevel,
		Stdout: true,
		// File:   "log/indexer/indexer.log",
		File: "",
	})
	if err != nil {
		panic(err)
	}
	defer sync()
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
	)
	if err != nil {
		logger.Error("migrate db error", zap.Error(err))
		return
	}

	// prepare service
	accountRepo := adapter.NewAccountRepo(db)
	inscriptionRepo := adapter.NewInscriptionRepo(db)

	svc := service.NewService(
		accountRepo,
		inscriptionRepo,
	)

	// init flow client
	flowClient, err := client.New("access.mainnet.nodes.onflow.org:9000", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}

	// scan
	maxBlockQuery := uint64(250)
	startBlock := uint64(68277132) // freeflow deployment block
	latestBlock, err := flowClient.GetLatestBlock(context.Background(), true)
	if err != nil {
		panic(err)
	}
	endBlock := latestBlock.Height

	for i := startBlock; i <= endBlock; i += maxBlockQuery - 1 {
		if i > endBlock {
			i = endBlock
		}
		scanRangeEvents(i, i+maxBlockQuery-1, flowClient, logger, svc)
	}

	// for i := startBlock; i <= endBlock; i++ {
	// 	getBlockTxs(i, flowClient, logger, svc)
	// }
}

func scanRangeEvents(startBlock, endBlock uint64, flowClient *client.Client, logger *zap.Logger, svc service.Service) {
	freeflowDepositEventType := "A.88dd257fcf26d3cc.Inscription.Deposit"
	bes, err := flowClient.GetEventsForHeightRange(context.Background(),
		client.EventRangeQuery{
			Type:        freeflowDepositEventType,
			StartHeight: startBlock,
			EndHeight:   endBlock,
		})
	if err != nil {
		logger.Error("GetEventsForHeightRange", zap.Error(err))
		return
	}

	for _, be := range bes {
		// logger.Info("BlockEvent", zap.Uint64("BlockHeight", be.Height))
		for _, e := range be.Events {
			if e.Type != freeflowDepositEventType {
				continue
			}
			// logger.Info("Event", zap.String("Type", e.Type))
			// logger.Info("Event", zap.String("TransactionID", e.TransactionID.String()))
			// logger.Info("Event", zap.String("TransactionIndex", fmt.Sprintf("%d", e.TransactionIndex)))
			// logger.Info("Event", zap.String("EventIndex", fmt.Sprintf("%d", e.EventIndex)))

			flowEvent := flowUtils.FreeflowDeposit(e)
			// logger.Info("Event", zap.Uint64("ID", flowEvent.ID()))
			// logger.Info("Event", zap.String("Address", fmt.Sprintf("%x", flowEvent.Address())))

			err := svc.UpdateBalance(context.Background(), "freeflow", fmt.Sprintf("%x", flowEvent.Address()), true)
			if err != nil {
				logger.Error("UpdateBalance", zap.Error(err))
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
