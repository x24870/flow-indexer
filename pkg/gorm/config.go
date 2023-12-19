package gorm

import "time"

type Config struct {
	DSN             string
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime time.Duration
	SingularTable   bool
}

func padDefault(config Config) Config {
	if config.MaxIdleConns <= 0 {
		config.MaxIdleConns = 2
	}
	if config.MaxOpenConns <= 0 {
		config.MaxOpenConns = 2
	}
	if config.ConnMaxLifetime <= 0 {
		config.ConnMaxLifetime = 10 * time.Minute
	}

	return config
}
