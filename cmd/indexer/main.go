package main

import (
	"context"
	"flow-indexer/pkg/log"
	"fmt"

	flowUtils "flow-indexer/pkg/flow"

	"github.com/onflow/flow-go-sdk/client"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
)

func main() {
	// init logger
	sync, err := log.Init(log.Config{
		Name:   "cart-backend.api",
		Level:  zapcore.DebugLevel,
		Stdout: true,
		// File:   "log/cart-backend/api.log",
		File: "",
	})
	if err != nil {
		panic(err)
	}
	defer sync()
	logger := zap.L()

	// init flow client
	flowClient, err := client.New("access.mainnet.nodes.onflow.org:9000", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}

	// scan
	maxBlockQuery := uint64(250)
	startBlock := uint64(68277132) // freeflow deployment block
	endBlock := startBlock + maxBlockQuery - 1

	scanRangeEvents(startBlock, endBlock, flowClient, logger)
}

func scanRangeEvents(startBlock, endBlock uint64, flowClient *client.Client, logger *zap.Logger) {
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
			logger.Info("Event", zap.String("Type", e.Type))
			logger.Info("Event", zap.String("TransactionID", e.TransactionID.String()))
			logger.Info("Event", zap.String("TransactionIndex", fmt.Sprintf("%d", e.TransactionIndex)))
			logger.Info("Event", zap.String("EventIndex", fmt.Sprintf("%d", e.EventIndex)))

			flowEvent := flowUtils.FreeflowDeposit(e)
			logger.Info("Event", zap.Uint64("ID", flowEvent.ID()))
			logger.Info("Event", zap.String("Address", fmt.Sprintf("%x", flowEvent.Address())))
			// logger.Info("Event", zap.Uint64("EventIndex", flowEvent.EventIndex()))
		}
	}
}

func getBlockTxs(blockNum uint64, flowClient *client.Client, logger *zap.Logger) {
	ctx := context.Background()
	block, err := flowClient.GetBlockByHeight(ctx, blockNum)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Block ID: %s\n", block.ID)
	for _, c := range block.CollectionGuarantees {
		col, err := flowClient.GetCollection(ctx, c.CollectionID)
		if err != nil {
			logger.Error("GetCollection", zap.Error(err))
		}

		for _, tID := range col.TransactionIDs {
			logger.Info("TransactionID", zap.String("ID", tID.String()))
			tx, err := flowClient.GetTransactionResult(ctx, tID)
			if err != nil {
				logger.Error("GetTransaction", zap.Error(err))
			}
			logger.Info("Transaction", zap.String("ID", tx.TransactionID.String()))
		}
	}

}
