package flow

import (
	"context"
	"flow-indexer/internal/service"
	"fmt"
	"sync"

	"github.com/onflow/flow-go-sdk/client"
	"go.uber.org/zap"
)

type BlockRange struct {
	StartBlock uint64
	EndBlock   uint64
}

func GetBlockRanges(startBlock, endBlock, thread uint64) []BlockRange {
	totalBlocks := endBlock - startBlock + 1
	groupSize := totalBlocks / thread
	if groupSize == 0 {
		groupSize = 1
	}

	var blockRanges []BlockRange
	for i := startBlock; i <= endBlock; i += groupSize {
		end := i + groupSize - 1
		if end > endBlock {
			end = endBlock
		}

		blockRanges = append(blockRanges, BlockRange{
			StartBlock: i,
			EndBlock:   end,
		})

		if end == endBlock {
			break
		}
	}

	return blockRanges
}

func ScanRangeEvents(
	startBlock, endBlock, maxBlockQuery uint64,
	flowClient *client.Client,
	logger *zap.Logger,
	svc service.Service,
	eventType string,
	wg *sync.WaitGroup,
) {
	go func() {
		for i := startBlock; i <= endBlock; i += maxBlockQuery {
			end := i + maxBlockQuery - 1
			if end > endBlock {
				end = endBlock
			}

			ScanBatchEvents(i, end, flowClient, logger, svc, eventType)
		}
		defer wg.Done()
	}()
}

func ScanBatchEvents(
	startBlock, endBlock uint64, flowClient *client.Client, logger *zap.Logger, svc service.Service, eventType string,
) {
	bes, err := flowClient.GetEventsForHeightRange(context.Background(),
		client.EventRangeQuery{
			Type:        eventType,
			StartHeight: startBlock,
			EndHeight:   endBlock,
		})
	if err != nil {
		logger.Error("GetEventsForHeightRange", zap.Error(fmt.Errorf("range %v - %v: %s", startBlock, endBlock, err)))
		return
	}

	for _, be := range bes {
		logger.Debug("BlockEvent", zap.Uint64("BlockHeight", be.Height))
		for _, e := range be.Events {
			if e.Type != eventType {
				continue
			}
			logger.Debug("Event", zap.String("Type", e.Type))
			logger.Debug("Event", zap.String("TransactionID", e.TransactionID.String()))
			logger.Debug("Event", zap.String("TransactionIndex", fmt.Sprintf("%d", e.TransactionIndex)))
			logger.Debug("Event", zap.String("EventIndex", fmt.Sprintf("%d", e.EventIndex)))

			flowEvent := FreeflowDeposit(e)
			logger.Debug("Event", zap.Uint64("ID", flowEvent.ID()))
			logger.Debug("Event", zap.String("Address", fmt.Sprintf("%x", flowEvent.Address())))

			err := svc.CreateFlowEvent(context.Background(), fmt.Sprintf("%x", flowEvent.Address()), e.Type, be.Height)
			if err != nil {
				logger.Error("CreateFlowEvent", zap.Error(err))
				return
			}

			err = svc.UpdateBalance(context.Background(), "freeflow", fmt.Sprintf("%x", flowEvent.Address()), eventType == FreeflowDepositEventType)
			if err != nil {
				logger.Error("UpdateBalance", zap.Error(
					fmt.Errorf(
						"address: %x, height: %v, range %v - %v: %e", flowEvent.Address(), be.Height, startBlock, endBlock, err,
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
	logger.Debug("Block", zap.Uint64("BlockHeight", blockNum))
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
			logger.Debug("TransactionID", zap.String("ID", tID.String()))
			tx, err := flowClient.GetTransactionResult(ctx, tID)
			if err != nil {
				logger.Error("GetTransaction", zap.Error(err))
			}
			logger.Debug("Transaction", zap.String("ID", tx.TransactionID.String()))
			logger.Debug("Transaction", zap.String("Status", tx.Status.String()))

			for _, e := range tx.Events {
				logger.Debug("Event", zap.String("Type", e.Type))
				if e.Type != "A.88dd257fcf26d3cc.Inscription.Withdraw" {
					continue
				}
				logger.Debug("Event", zap.String("TransactionID", e.TransactionID.String()))
				logger.Debug("Event", zap.String("TransactionIndex", fmt.Sprintf("%d", e.TransactionIndex)))
				logger.Debug("Event", zap.String("EventIndex", fmt.Sprintf("%d", e.EventIndex)))

				flowEvent := FreeflowDeposit(e)
				logger.Debug("Event", zap.Uint64("ID", flowEvent.ID()))
				logger.Debug("Event", zap.String("Address", fmt.Sprintf("%x", flowEvent.Address())))

				err := svc.UpdateBalance(ctx, "freeflow", fmt.Sprintf("%x", flowEvent.Address()), e.Type == FreeflowDepositEventType)
				if err != nil {
					logger.Error("UpdateBalance", zap.Error(err))
					return
				}
			}
		}
	}

}
