package commands

import (
	"fmt"
	"strconv"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/handler"
	"github.com/disgoorg/snowflake/v2"

	"github.com/KittyBot-Org/KittyBotGo/db/.gen/kittybot-go/public/model"
	"github.com/KittyBot-Org/KittyBotGo/dbot"
)

func Report(b *dbot.Bot) handler.Command {
	return handler.Command{
		Create: discord.MessageCommandCreate{
			Name: "report",
		},
		CommandHandlers: map[string]handler.CommandHandler{
			"": reportHandler(b),
		},
	}
}

func reportHandler(b *dbot.Bot) handler.CommandHandler {
	return func(e *events.ApplicationCommandInteractionCreate) error {
		data := e.MessageCommandInteractionData()

		msg := data.TargetMessage()
		if msg.Author.ID == e.User().ID {
			return e.CreateMessage(discord.MessageCreate{
				Content: "You cannot report your own messages.",
				Flags:   discord.MessageFlagEphemeral,
			})
		}
		if msg.Author.Bot {
			return e.CreateMessage(discord.MessageCreate{
				Content: "You cannot report bot messages.",
				Flags:   discord.MessageFlagEphemeral,
			})
		}

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

		reportID, err := b.DB.Reports().Create(msg.Author.ID, *e.GuildID(), msg.Content, time.Now(), msg.ID, msg.ChannelID)
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

		messageURL := fmt.Sprintf("https://discord.com/channels/%s/%s/%s", *e.GuildID(), msg.ChannelID, msg.ID)

		return CreateReport(b, settings, reportID,
			fmt.Sprintf("%s(%s)'s [message](%s) has been reported by %s(%s).\nCreated a new report with the id `%d`", msg.Author.Tag(), msg.Author.Mention(), messageURL, e.User().Tag(), e.User().Mention(), reportID),
			discord.Embed{
				Author: &discord.EmbedAuthor{
					Name:    msg.Author.Username,
					URL:     messageURL,
					IconURL: msg.Author.EffectiveAvatarURL(),
				},
				Description: msg.Content,
				Timestamp:   &msg.CreatedAt,
			},
		)
	}
}

func CreateReport(b *dbot.Bot, settings model.GuildSettings, reportID int32, content string, embed discord.Embed) error {
	client := b.ReportLogWebhookMap.Get(snowflake.MustParse(settings.ModerationLogWebhookID), settings.ModerationLogWebhookToken)
	_, err := client.CreateMessage(discord.WebhookMessageCreate{
		Content: content,
		Embeds: []discord.Embed{
			embed,
		},
		Components: []discord.ContainerComponent{
			discord.ActionRowComponent{
				discord.ButtonComponent{
					Style:    discord.ButtonStyleSuccess,
					Label:    "Confirm",
					CustomID: fmt.Sprintf("handler:report-confirm:%d", reportID),
				},
				discord.ButtonComponent{
					Style:    discord.ButtonStyleDanger,
					Label:    "Delete",
					CustomID: fmt.Sprintf("handler:report-delete:%d", reportID),
				},
			},
		},
	})
	return err
}

func parseReportID(arg string) (int32, error) {
	id, err := strconv.ParseInt(arg, 10, 32)
	if err != nil {
		return 0, err
	}
	return int32(id), nil
}
