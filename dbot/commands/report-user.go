package commands

import (
	"fmt"
	"time"

	"github.com/KittyBot-Org/KittyBotGo/dbot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/handler"
	"github.com/disgoorg/json"
)

func ReportUser(b *dbot.Bot) handler.Command {
	return handler.Command{
		Create: discord.SlashCommandCreate{
			Name:        "report-user",
			Description: "Report a user for breaking the rules.",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionUser{
					Name:        "user",
					Description: "The user to report.",
					Required:    true,
				},
				discord.ApplicationCommandOptionString{
					Name:        "reason",
					Description: "The reason for the report.",
					Required:    true,
				},
			},
		},
		CommandHandlers: map[string]handler.CommandHandler{
			"": reportUserHandler(b),
		},
	}
}

func reportUserHandler(b *dbot.Bot) handler.CommandHandler {
	return func(e *events.ApplicationCommandInteractionCreate) error {
		data := e.SlashCommandInteractionData()
		user := data.User("user")
		reason := data.String("reason")

		settings, err := b.DB.GuildSettings().Get(*e.GuildID())
		if err != nil {
			b.Logger.Errorf("Failed to get guild settings: %s", err)
			return e.CreateMessage(discord.MessageCreate{
				Content: "Failed to get guild settings, please reach out to a bot developer.",
				Flags:   discord.MessageFlagEphemeral,
			})
		}
		if settings.ModerationLogWebhookID == "0" {
			return e.CreateMessage(discord.MessageCreate{
				Content: "Moderation is not enabled on this server, please reach to a moderator.",
				Flags:   discord.MessageFlagEphemeral,
			})
		}

		reportID, err := b.DB.Reports().Create(user.ID, *e.GuildID(), reason, time.Now(), 0, 0)
		if err != nil {
			b.Logger.Errorf("Failed to create report: %s", err)
			return e.CreateMessage(discord.MessageCreate{
				Content: "Failed to create report, please reach out to a moderator.",
				Flags:   discord.MessageFlagEphemeral,
			})
		}

		err = e.CreateMessage(discord.MessageCreate{
			Content: "Successfully reported.",
			Flags:   discord.MessageFlagEphemeral,
		})
		if err != nil {
			b.Logger.Errorf("Failed to send report confirmation message: %s", err)
		}

		return CreateReport(b, settings, reportID,
			fmt.Sprintf("%s(%s) has been reported by %s(%s).\nCreated a new report with the id #`%d`", user.Tag(), user.Mention(), e.User().Tag(), e.User().Mention(), reportID),
			discord.Embed{
				Author: &discord.EmbedAuthor{
					Name:    user.Username,
					IconURL: user.EffectiveAvatarURL(),
				},
				Description: "Reason:\n" + reason,
				Timestamp:   json.Ptr(e.ID().Time()),
			},
		)
	}
}
