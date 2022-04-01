package db

import (
	"context"
	"database/sql"

	"github.com/uptrace/bun/extra/bundebug"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
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
	db := bun.NewDB(sqlDB, pgdialect.New())
	db.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(devMode)))
	if shouldSyncDBTables {
		if err := db.ResetModel(context.TODO(), (*VoterModel)(nil)); err != nil {
			return nil, err
		}
		if err := db.ResetModel(context.TODO(), (*GuildSettings)(nil)); err != nil {
			return nil, err
		}
		if err := db.ResetModel(context.TODO(), (*Tag)(nil)); err != nil {
			return nil, err
		}
		if err := db.ResetModel(context.TODO(), (*MusicPlayer)(nil)); err != nil {
			return nil, err
		}
		if err := db.ResetModel(context.TODO(), (*PlayHistory)(nil)); err != nil {
			return nil, err
		}
		if err := db.ResetModel(context.TODO(), (*LikedSong)(nil)); err != nil {
			return nil, err
		}
	}
	return db, nil
}
