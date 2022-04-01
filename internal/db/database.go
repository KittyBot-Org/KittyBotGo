package db

import (
	"database/sql"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/driver/pgdriver"
)

type DatabaseConfig struct {
	Address  string `json:"address"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"db_name"`
}

func SetupDatabase(config DatabaseConfig, shouldSyncDBTables bool, devMode bool) (*bun.DB, error) {
	sqlDB := sql.OpenDB(pgdriver.NewConnector(
		pgdriver.WithAddr(config.Address),
		pgdriver.WithUser(config.User),
		pgdriver.WithPassword(config.Password),
		pgdriver.WithDatabase(config.DBName),
		pgdriver.WithInsecure(true),
	))

	return db, nil
}
