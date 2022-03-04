package types

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/KittyBot-Org/KittyBotGo/internal/models"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

func (b *Bot) SetupDatabase(shouldSyncDBTables bool) error {
	sqlDB := sql.OpenDB(pgdriver.NewConnector(
		pgdriver.WithAddr(fmt.Sprintf("%s:%s", b.Config.Database.Host, b.Config.Database.Port)),
		pgdriver.WithUser(b.Config.Database.User),
		pgdriver.WithPassword(b.Config.Database.Password),
		pgdriver.WithDatabase(b.Config.Database.DBName),
		pgdriver.WithInsecure(true),
	))
	b.DB = bun.NewDB(sqlDB, pgdialect.New())

	if shouldSyncDBTables {
		if err := b.DB.ResetModel(context.TODO(), (*models.Tag)(nil)); err != nil {
			return err
		}
		if err := b.DB.ResetModel(context.TODO(), (*models.MusicPlayer)(nil)); err != nil {
			return err
		}
		if err := b.DB.ResetModel(context.TODO(), (*models.PlayHistory)(nil)); err != nil {
			return err
		}
	}
	return nil
}
