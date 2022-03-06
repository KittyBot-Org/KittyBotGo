package types

import (
	"context"
	"github.com/KittyBot-Org/KittyBotGo/internal/cache"
	"time"

	"github.com/DisgoOrg/snowflake"
	"github.com/KittyBot-Org/KittyBotGo/internal/models"
)

func NewPlayHistoryCache(bot *Bot) *cache.Cache[snowflake.Snowflake, []models.PlayHistory] {
	return cache.New[snowflake.Snowflake, []models.PlayHistory](5*time.Minute,
		func(userID snowflake.Snowflake) (history []models.PlayHistory) {
			if err := bot.DB.NewSelect().Model(&history).Where("user_id = ?", userID).Scan(context.TODO()); err != nil {
				bot.Logger.Error("Failed to get music history entries: ", err)
				return nil
			}
			return history
		},
		func(userID snowflake.Snowflake, history []models.PlayHistory) {
			if _, err := bot.DB.NewInsert().Model(&history).On("CONFLICT (id) DO UPDATE").Exec(context.TODO()); err != nil {
				bot.Logger.Error("Error adding music history entry: ", err)
			}
		})
}
