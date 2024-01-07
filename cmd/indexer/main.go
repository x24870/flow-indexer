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
		Level:  zapcore.InfoLevel,
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
	thread := 15
	maxBlockQuery := uint64(250) - 1
	startBlock := uint64(68277132) // freeflow deployment block
	latestBlock, err := flowClient.GetLatestBlock(context.Background(), true)
	if err != nil {
		panic(err)
	}
	endBlock := latestBlock.Height
	// endBlock := uint64(69106534)

	logger.Info("start scan", zap.Uint64("startBlock", startBlock), zap.Uint64("endBlock", endBlock))
	blockRanges := flowUtils.GetBlockRanges(startBlock, endBlock, uint64(thread))
	var wg sync.WaitGroup
	for _, blockRange := range blockRanges {
		wg.Add(1)
		logger.Info("scan range", zap.Uint64("startBlock", blockRange.StartBlock), zap.Uint64("endBlock", blockRange.EndBlock))
		flowUtils.ScanRangeEvents(
			blockRange.StartBlock,
			blockRange.EndBlock,
			maxBlockQuery,
			flowClient,
			logger,
			svc,
			flowUtils.FreeflowDepositEventType, // FreeflowDepositEventType
			&wg,
		)
	}

	wg.Wait()
}
