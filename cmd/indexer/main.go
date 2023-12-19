package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

func main() {
	// sleep 1s to wait for db
	time.Sleep(1 * time.Second)

	// load env
	err := LoadEnvConfig(".env")
	if err != nil {
		panic(fmt.Errorf("failed to load config: %v", err))
	}

	// init logger
	// sync, err := log.Init(log.Config{
	// 	Name:   "cart-backend.api",
	// 	Level:  zapcore.DebugLevel,
	// 	Stdout: true,
	// 	// File:   "log/cart-backend/api.log",
	// 	File: "",
	// })
	// if err != nil {
	// 	panic(err)
	// }
	// defer sync()
	// logger := zap.L()

	// prepare context
	// ctx := app.GraceCtx(context.Background())

	// init db
	// dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
	// 	os.Getenv("DB_USERNAME"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_NAME"))
	// logger.Info("dsn", zap.String("dsn", dsn))

	// db, err := gormpkg.NewGormPostgresConn(
	// 	gormpkg.Config{
	// 		DSN:             dsn,
	// 		MaxIdleConns:    2,
	// 		MaxOpenConns:    2,
	// 		ConnMaxLifetime: 10 * time.Minute,
	// 		SingularTable:   true,
	// 	},
	// )
	// if err != nil {
	// 	logger.Error("connect to database error", zap.Error(err))
	// 	return
	// }

	// // create extension
	// db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"")

	// // migrate db
	// err = db.AutoMigrate(
	// 	&account.Account{},
	// 	&t.TxRecord{},
	// 	&t.Operation{},
	// 	&t.Intent{},
	// )
	// if err != nil {
	// 	logger.Error("migrate db error", zap.Error(err))
	// 	return
	// }

	// prepare service
	// svc := service.NewService()

}

func LoadEnvConfig(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "=", 2)

		// if starts with #, skip
		if strings.HasPrefix(line, "#") {
			continue
		}

		if len(parts) != 2 {
			return fmt.Errorf("%s invalid line: %s", filename, line)
		}
		key := parts[0]
		value := parts[1]
		os.Setenv(key, value) // set as environment variable
	}

	return scanner.Err()
}
