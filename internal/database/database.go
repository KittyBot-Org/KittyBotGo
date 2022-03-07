package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/uptrace/bun/extra/bundebug"

	"github.com/KittyBot-Org/KittyBotGo/internal/models"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

func SetupDatabase(config Config, shouldSyncDBTables bool, devMode bool) (*bun.DB, error) {
	sqlDB := sql.OpenDB(pgdriver.NewConnector(
		pgdriver.WithAddr(fmt.Sprintf("%s:%s", config.Host, config.Port)),
		pgdriver.WithUser(config.User),
		pgdriver.WithPassword(config.Password),
		pgdriver.WithDatabase(config.DBName),
		pgdriver.WithInsecure(true),
	))
	db := bun.NewDB(sqlDB, pgdialect.New())
	db.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(devMode)))
	if shouldSyncDBTables {
		if err := db.ResetModel(context.TODO(), (*models.GuildSettings)(nil)); err != nil {
			return nil, err
		}
		if err := db.ResetModel(context.TODO(), (*models.Tag)(nil)); err != nil {
			return nil, err
		}
		if err := db.ResetModel(context.TODO(), (*models.MusicPlayer)(nil)); err != nil {
			return nil, err
		}
		if err := db.ResetModel(context.TODO(), (*models.PlayHistory)(nil)); err != nil {
			return nil, err
		}
		if err := db.ResetModel(context.TODO(), (*models.LikedSong)(nil)); err != nil {
			return nil, err
		}
	}
	return db, nil
}
