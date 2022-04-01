package dbot

import (
	"context"

	"github.com/KittyBot-Org/KittyBotGo/internal/db"
	"github.com/disgoorg/snowflake"
)

func (b *Bot) AddPlayHistory(userID snowflake.Snowflake, title string, query string) {
	entry := db.PlayHistory{
		UserID: userID,
		Title:  title,
		Query:  query,
	}
	if _, err := b.DB.NewInsert().Model(&entry).On("CONFLICT (user_id, title) DO UPDATE").Exec(context.TODO()); err != nil {
		b.Logger.Error("Error adding music history entry: ", err)
	}
	// TODO: fix this lol
	if _, err := b.DB.NewDelete().Model((*db.PlayHistory)(nil)).Where("last_used_at <= (?)",
		b.DB.NewSelect().Column("last_used_at").Table("play_histories").Where("user_id = ?", userID).Order("last_used_at DESC").Limit(1).Offset(b.Config.PlayHistoryCacheSize),
	).Exec(context.TODO()); err != nil {
		b.Logger.Error("Error deleting music history entry: ", err)
	}
}
