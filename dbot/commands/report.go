package commands

import (
	"fmt"
	"strconv"
	"time"

	"github.com/KittyBot-Org/KittyBotGo/dbot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/webhook"
	"github.com/disgoorg/snowflake/v2"
	"golang.org/x/text/message"
)

var Report = dbot.Command{
	Create: discord.MessageCommandCreate{
		CommandName:  "report",
		DMPermission: false,
	},
	CommandHandler: map[string]dbot.CommandHandler{
		"": reportHandler,
	},
	ComponentHandler: map[string]dbot.ComponentHandler{
		"action":  reportActionHandler,
		"confirm": reportConfirmHandler,
		"delete":  reportDeleteHandler,
	},
}

func reportHandler(b *dbot.Bot, p *message.Printer, e *events.ApplicationCommandInteractionCreate) error {
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

	client, ok := b.ReportLogWebhooks[*e.GuildID()]
	if !ok {
		settings, err := b.DB.GuildSettings().Get(*e.GuildID())
		if err != nil {
			return err
		}
		if settings.ModerationLogWebhookID == "" {
			return e.CreateMessage(discord.MessageCreate{
				Content: "Moderation is not enabled on this server, please reach to a moderator.",
				Flags:   discord.MessageFlagEphemeral,
			})
		}
		b.ReportLogWebhooks[*e.GuildID()] = webhook.New(snowflake.MustParse(settings.ModerationLogWebhookID), settings.ModerationLogWebhookToken)
	}

	incidentID, err := b.DB.Reports().Create(msg.Author.ID, *e.GuildID(), msg.Content, time.Now(), msg.ID)
	if err != nil {
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

	_, err = client.CreateMessage(discord.WebhookMessageCreate{
		Content: fmt.Sprintf("%s(%s) has reported a message from %s(%s)\nCreated a new incident with the id `%d`", e.User().Tag(), e.User().Mention(), msg.Author.Tag(), msg.Author.Mention(), incidentID),
		Embeds: []discord.Embed{
			{
				Author: &discord.EmbedAuthor{
					Name:    msg.Author.Username,
					URL:     fmt.Sprintf("https://discord.com/channels/%s/%s/%s", *e.GuildID(), msg.ChannelID, msg.ID),
					IconURL: msg.Author.EffectiveAvatarURL(),
				},
				Description: msg.Content,
				Timestamp:   &msg.CreatedAt,
			},
		},
		Components: []discord.ContainerComponent{
			discord.ActionRowComponent{
				discord.ButtonComponent{
					Style:    discord.ButtonStyleSuccess,
					Label:    "Confirm",
					CustomID: discord.CustomID(fmt.Sprintf("report:confirm:%d", incidentID)),
				},
				discord.ButtonComponent{
					Style:    discord.ButtonStyleDanger,
					Label:    "Delete",
					CustomID: discord.CustomID(fmt.Sprintf("report:delete:%d", incidentID)),
				},
			},
		},
	})
	return err
}

func reportConfirmHandler(b *dbot.Bot, args []string, p *message.Printer, e *events.ComponentInteractionCreate) error {
	var incidentID int32
	if rawIncidentID, err := strconv.ParseInt(args[0], 10, 32); err != nil {
		return err
	} else {
		incidentID = int32(rawIncidentID)
	}

	report, err := b.DB.Reports().Get(incidentID)
	if err != nil {
		b.Logger.Errorf("Failed to get report: %s", err)
		return e.CreateMessage(discord.MessageCreate{
			Content: "Failed to confirm report, please reach out to a bot developer.",
			Flags:   discord.MessageFlagEphemeral,
		})
	}
	if err = b.DB.Reports().Confirm(incidentID); err != nil {
		b.Logger.Errorf("Failed to confirm report: %s", err)
		return e.CreateMessage(discord.MessageCreate{
			Content: "Failed to confirm report, please reach out to a bot developer.",
			Flags:   discord.MessageFlagEphemeral,
		})
	}

	return e.UpdateMessage(discord.MessageUpdate{
		Components: &[]discord.ContainerComponent{
			discord.ActionRowComponent{
				discord.SelectMenuComponent{
					CustomID: discord.CustomID(fmt.Sprintf("report:action:%d", incidentID)),
					Options: []discord.SelectMenuOption{
						{
							Label: "Delete Message",
							Value: fmt.Sprintf("delete:%d", report.MessageID),
						},
						{
							Label: "Timeout User for 1 hour",
							Value: "timeout:1",
							Emoji: &discord.ComponentEmoji{
								Name: "🚫",
							},
						},
						{
							Label: "Timeout User for 1 day",
							Value: "timeout:24",
							Emoji: &discord.ComponentEmoji{
								Name: "🚫",
							},
						},
						{
							Label: "Kick User",
							Value: "kick",
							Emoji: &discord.ComponentEmoji{
								Name: "👞",
							},
						},
						{
							Label: "Ban User",
							Value: "ban",
							Emoji: &discord.ComponentEmoji{
								Name: "🔨",
							},
						},
					},
				},
			},
		},
	})
}

func reportDeleteHandler(b *dbot.Bot, args []string, p *message.Printer, e *events.ComponentInteractionCreate) error {
	var incidentID int32
	if rawIncidentID, err := strconv.ParseInt(args[0], 10, 32); err != nil {
		return err
	} else {
		incidentID = int32(rawIncidentID)
	}

	if err := b.DB.Reports().Delete(incidentID); err != nil {
		b.Logger.Errorf("Failed to delete report: %s", err)
		return e.CreateMessage(discord.MessageCreate{
			Content: "Failed to delete report, please reach out to a bot developer.",
			Flags:   discord.MessageFlagEphemeral,
		})
	}

	if err := e.DeferUpdateMessage(); err != nil {
		b.Logger.Errorf("Failed to defer update message: %s", err)
		return err
	}

	if err := e.Client().Rest().DeleteInteractionResponse(e.ApplicationID(), e.Token()); err != nil {
		b.Logger.Errorf("Failed to delete interaction response: %s", err)
		return err
	}

	_, err := b.Client.Rest().CreateFollowupMessage(e.ApplicationID(), e.Token(), discord.MessageCreate{
		Content: "Report deleted.",
		Flags:   discord.MessageFlagEphemeral,
	})
	return err
}

func reportActionHandler(b *dbot.Bot, args []string, p *message.Printer, e *events.ComponentInteractionCreate) error {
	return nil
}
