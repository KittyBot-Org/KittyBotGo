package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/kittybot-org/kittybotgo/service/bot/db/dbsqlc"
	"github.com/kittybot-org/kittybotgo/service/bot/res"
)

var notificationCommand = discord.SlashCommandCreate{
	Name:        "notification",
	Description: "Remind yourself later",
	Options: []discord.ApplicationCommandOption{
		discord.ApplicationCommandOptionSubCommand{Name: "create", Description: "Creates a notification", Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionString{Name: "in", Description: "How long to wait, for example 1h30m", Required: true},
			discord.ApplicationCommandOptionString{Name: "content", Description: "What to remind you about", Required: true},
		}},
		discord.ApplicationCommandOptionSubCommand{Name: "delete", Description: "Deletes one of your notifications", Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionInt{Name: "id", Description: "The notification id", Required: true},
		}},
		discord.ApplicationCommandOptionSubCommand{Name: "list", Description: "Lists your notifications"},
	},
}

func (c *commands) OnNotificationCreate(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	delay, err := time.ParseDuration(data.String("in"))
	if err != nil || delay <= 0 {
		return e.CreateMessage(res.CreateError("Give a duration like 10m, 2h, or 1h30m"))
	}
	err = c.Database.CreateNotification(e.Ctx, dbsqlc.CreateNotificationParams{
		GuildID:          int64(*e.GuildID()),
		UserID:           int64(e.User().ID),
		ChannelID:        int64(e.Channel().ID()),
		Content:          data.String("content"),
		NotificationTime: pgtype.Timestamp{Time: time.Now().Add(delay), Valid: true},
	})
	if err != nil {
		return e.CreateMessage(res.CreateErr("Failed to create that notification", err))
	}
	return e.CreateMessage(res.Createf("I will remind you in %s", delay))
}

func (c *commands) OnNotificationDelete(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	if err := c.Database.DeleteNotification(e.Ctx, dbsqlc.DeleteNotificationParams{
		ID:     int64(data.Int("id")),
		UserID: int64(e.User().ID),
	}); err != nil {
		return e.CreateMessage(res.CreateErr("Failed to delete that notification", err))
	}
	return e.CreateMessage(res.Create("Deleted the notification"))
}

func (c *commands) OnNotificationList(_ discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	notifications, err := c.Database.ListNotifications(e.Ctx, int64(e.User().ID))
	if err != nil {
		return e.CreateMessage(res.CreateErr("Failed to list notifications", err))
	}
	if len(notifications) == 0 {
		return e.CreateMessage(res.Create("You have no notifications"))
	}
	var lines []string
	for _, notification := range notifications {
		lines = append(lines, fmt.Sprintf("`%d` %s - %s", notification.ID, notification.NotificationTime.Time.Format(time.RFC822), notification.Content))
	}
	return e.CreateMessage(res.Create(strings.Join(lines, "\n")))
}
