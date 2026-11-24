package bot

import (
	"context"
	"log/slog"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/snowflake/v2"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/topi314/tint"

	"github.com/kittybot-org/kittybotgo/service/bot/db/dbsqlc"
)

func (b *Bot) scheduleNotifications() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-b.ctx.Done():
			return
		case <-ticker.C:
			b.sendDueNotifications()
		}
	}
}

func (b *Bot) sendDueNotifications() {
	notifications, err := b.Database.DueNotifications(b.ctx, pgtype.Timestamp{Time: time.Now(), Valid: true})
	if err != nil {
		slog.Error("failed to load notifications", tint.Err(err))
		return
	}
	for _, notification := range notifications {
		_, err := b.Discord.Rest.CreateMessage(snowflake.ID(notification.ChannelID), discord.MessageCreate{
			Content: discord.UserMention(snowflake.ID(notification.UserID)),
			Embeds: []discord.Embed{{
				Title:       "Notification",
				Description: notification.Content,
				Color:       0x5c5fea,
			}},
		})
		if err != nil {
			slog.Error("failed to send notification", tint.Err(err))
		}
		if err := b.Database.DeleteNotification(context.Background(), dbsqlc.DeleteNotificationParams{
			ID:     notification.ID,
			UserID: notification.UserID,
		}); err != nil {
			slog.Error("failed to delete notification", tint.Err(err))
		}
	}
}
