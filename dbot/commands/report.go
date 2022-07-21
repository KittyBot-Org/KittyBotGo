package commands

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/KittyBot-Org/KittyBotGo/dbot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/json"
	"github.com/disgoorg/disgo/rest"
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

	settings, err := b.DB.GuildSettings().Get(*e.GuildID())
	if settings.ModerationLogWebhookID == "" {
		return e.CreateMessage(discord.MessageCreate{
			Content: "Moderation is not enabled on this server, please reach to a moderator.",
			Flags:   discord.MessageFlagEphemeral,
		})
	}

	client := b.ReportLogWebhookMap.Get(snowflake.MustParse(settings.ModerationLogWebhookID), settings.ModerationLogWebhookToken)

	incidentID, err := b.DB.Reports().Create(msg.Author.ID, *e.GuildID(), msg.Content, time.Now(), msg.ID, msg.ChannelID)
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

func parseReportID(arg string) (int32, error) {
	id, err := strconv.ParseInt(arg, 10, 32)
	if err != nil {
		return 0, err
	}
	return int32(id), nil
}

func reportConfirmHandler(b *dbot.Bot, args []string, p *message.Printer, e *events.ComponentInteractionCreate) error {
	reportID, err := parseReportID(args[0])
	if err != nil {
		return err
	}

	if err = b.DB.Reports().Confirm(reportID); err != nil {
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
					CustomID:    discord.CustomID(fmt.Sprintf("report:action:%d", reportID)),
					Placeholder: "Select an action",
					MinValues:   json.NewPtr(1),
					MaxValues:   1,
					Options: []discord.SelectMenuOption{
						{
							Label: "Delete Message",
							Value: "delete",
							Emoji: &discord.ComponentEmoji{
								Name: "🗑",
							},
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
	reportID, err := parseReportID(args[0])
	if err != nil {
		return err
	}

	if err = b.DB.Reports().Delete(reportID); err != nil {
		b.Logger.Errorf("Failed to delete report: %s", err)
		return e.CreateMessage(discord.MessageCreate{
			Content: "Failed to delete report, please reach out to a bot developer.",
			Flags:   discord.MessageFlagEphemeral,
		})
	}

	if err = e.DeferUpdateMessage(); err != nil {
		b.Logger.Errorf("Failed to defer update message: %s", err)
		return err
	}

	if err = e.Client().Rest().DeleteInteractionResponse(e.ApplicationID(), e.Token()); err != nil {
		b.Logger.Errorf("Failed to delete interaction response: %s", err)
		return err
	}

	_, err = b.Client.Rest().CreateFollowupMessage(e.ApplicationID(), e.Token(), discord.MessageCreate{
		Content: "Report deleted.",
		Flags:   discord.MessageFlagEphemeral,
	})
	return err
}

func reportActionHandler(b *dbot.Bot, args []string, p *message.Printer, e *events.ComponentInteractionCreate) error {
	reportID, err := parseReportID(args[0])
	if err != nil {
		return err
	}

	report, err := b.DB.Reports().Get(reportID)
	if err != nil {
		b.Logger.Errorf("Failed to get report: %s", err)
		return e.CreateMessage(discord.MessageCreate{
			Content: "Failed to get report, please reach out to a bot developer.",
			Flags:   discord.MessageFlagEphemeral,
		})
	}

	reason := rest.WithReason(fmt.Sprintf("automod action by %s caused by report %d", e.User().Tag(), reportID))

	var content string
	value := e.SelectMenuInteractionData().Values[0]
	switch value {
	case "delete":
		if err = b.Client.Rest().DeleteMessage(snowflake.MustParse(report.ChannelID), snowflake.MustParse(report.MessageID), reason); err != nil {
			b.Logger.Errorf("Failed to delete message: %s", err)
			content = "Failed to delete message, please reach out to a bot developer."
		} else {
			content = "Message deleted."
		}

	case "timeout:1", "timeout:24":
		until := time.Now()
		if strings.HasSuffix(value, "24") {
			until = until.Add(24 * time.Hour)
		} else {
			until = until.Add(1 * time.Hour)
		}
		if _, err = b.Client.Rest().UpdateMember(*e.GuildID(), snowflake.MustParse(report.UserID), discord.MemberUpdate{
			CommunicationDisabledUntil: json.NewOptional(until),
		}, reason); err != nil {
			b.Logger.Errorf("Failed to update member: %s", err)
			content = "Failed to timeout user, please reach out to a bot developer."
		} else {
			content = "User timed out."
		}

	case "kick":
		if err = b.Client.Rest().RemoveMember(*e.GuildID(), snowflake.MustParse(report.UserID), reason); err != nil {
			b.Logger.Errorf("Failed to kick user: %s", err)
			content = "Failed to kick user, please reach out to a bot developer."
		} else {
			content = "User kicked."
		}

	case "ban":
		if err = b.Client.Rest().AddBan(*e.GuildID(), snowflake.MustParse(report.UserID), 0, reason); err != nil {
			b.Logger.Errorf("Failed to ban user: %s", err)
			content = "Failed to ban user, please reach out to a bot developer."
		} else {
			content = "User banned."
		}
	}
	return e.CreateMessage(discord.MessageCreate{
		Content: content,
		Flags:   discord.MessageFlagEphemeral,
	})
}
