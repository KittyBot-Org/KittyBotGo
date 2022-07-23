package listeners

import (
	"fmt"
	"time"

	"github.com/KittyBot-Org/KittyBotGo/dbot"
	"github.com/KittyBot-Org/KittyBotGo/dbot/commands"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/json"
	"github.com/disgoorg/snowflake/v2"
)

func Moderation(b *dbot.Bot) bot.EventListener {
	return bot.NewListenerFunc(func(e *events.AutoModerationActionExecution) {
		settings, err := b.DB.GuildSettings().Get(e.GuildID)
		if err != nil {
			b.Logger.Errorf("Failed to get guild settings: %s", err)
			return
		}
		if settings.ModerationLogWebhookID == "0" {
			return
		}

		var messageID snowflake.ID
		if e.MessageID != nil {
			messageID = *e.MessageID
		}
		var channelID snowflake.ID
		if e.ChannelID != nil {
			channelID = *e.ChannelID
		}

		reportID, err := b.DB.Reports().Create(e.UserID, e.GuildID, e.Content, time.Now(), messageID, channelID)
		if err != nil {
			b.Logger.Errorf("Failed to create report for automod execution: %s", err)
		}

		user, err := b.Client.Rest().GetUser(e.UserID)
		if err != nil {
			b.Logger.Errorf("Failed to get user: %s", err)
			return
		}

		var messageURL string
		if messageID != 0 && channelID != 0 {
			messageURL = fmt.Sprintf("https://discord.com/channels/%s/%s/%s", e.GuildID, channelID, messageID)
		}

		var content string
		if messageURL == "" {
			content = fmt.Sprintf("%s(%s)'s [message](%s) has triggered automod.\nCreated a new report with the id #`%d`", user.Tag(), user.Mention(), messageURL, reportID)
		} else {
			content = fmt.Sprintf("%s(%s)'s message has been blocked by automod.\nCreated a new report with the id #`%d`", user.Tag(), user.Mention(), reportID)
		}

		var fields []discord.EmbedField
		if e.MatchedContent != nil {
			fields = append(fields, discord.EmbedField{
				Name:  "Matched Content",
				Value: *e.MatchedContent,
			})
		}

		if e.MatchedKeywords != nil {
			fields = append(fields, discord.EmbedField{
				Name:  "Matched Keywords",
				Value: *e.MatchedKeywords,
			})
		}

		err = commands.CreateReport(b, settings, reportID,
			content,
			discord.Embed{
				Author: &discord.EmbedAuthor{
					Name:    user.Username,
					URL:     messageURL,
					IconURL: user.EffectiveAvatarURL(),
				},
				Description: e.Content,
				Fields:      fields,
				Timestamp:   json.NewPtr(time.Now()),
			},
		)
		if err != nil {
			b.Logger.Errorf("Failed to create automod report: %s", err)
		}
	})
}
