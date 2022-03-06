package types

import (
	"context"
	"github.com/DisgoOrg/snowflake"
	"github.com/KittyBot-Org/KittyBotGo/internal/models"
)

func (b *Bot) AddPlayHistory(userID snowflake.Snowflake, title string, query string) {
	entry := models.PlayHistory{
		UserID: userID,
		Title:  title,
		Query:  query,
	}
	if _, err := b.DB.NewInsert().Model(&entry).On("CONFLICT (id) DO UPDATE").Exec(context.TODO()); err != nil {
		b.Logger.Error("Error adding music history entry: ", err)
	}
	// TODO: fix this lol
	if _, err := b.DB.NewDelete().Model((*models.PlayHistory)(nil)).Where("id <= (?)",
		b.DB.NewSelect().Column("id").Table("play_histories").Where("user_id = ?", userID).Order("last_used_at DESC").Limit(1).Offset(4),
	).Exec(context.TODO()); err != nil {
		b.Logger.Error("Error deleting music history entry: ", err)
	}
}
