package commands

import (
	"fmt"
	"time"

	"github.com/KittyBot-Org/KittyBotGo/dbot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/json"
	"golang.org/x/text/message"
)

var ReportUser = dbot.Command{
	Create: discord.SlashCommandCreate{
		CommandName: "report-user",
		Description: "Report a user for breaking the rules.",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionUser{
				OptionName:  "user",
				Description: "The user to report.",
				Required:    true,
			},
			discord.ApplicationCommandOptionString{
				OptionName:  "reason",
				Description: "The reason for the report.",
				Required:    true,
			},
		},
	},
	CommandHandler: map[string]dbot.CommandHandler{
		"": reportUserHandler,
	},
}

func reportUserHandler(b *dbot.Bot, p *message.Printer, e *events.ApplicationCommandInteractionCreate) error {
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
			Timestamp:   json.NewPtr(e.ID().Time()),
		},
	)
}
