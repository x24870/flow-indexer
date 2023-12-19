package gorm

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

func NewGormPostgresConn(config Config) (*gorm.DB, error) {
	config = padDefault(config)
	db, err := gorm.Open(
		postgres.New(
			postgres.Config{
				DSN: config.DSN,
			}),
		&gorm.Config{
			Logger: logger.Default.LogMode(logger.Warn),
			NamingStrategy: schema.NamingStrategy{
				SingularTable: config.SingularTable,
			},
			// In order to use the dialect translated errors(like ErrDuplicatedKey),
			// enable the TranslateError flag when opening a db connection.
			TranslateError: true,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("gorm.Open error: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("db.DB error: %w", err)
	}

	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)

	return db, nil
}
