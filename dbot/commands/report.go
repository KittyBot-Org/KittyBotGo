package commands

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/KittyBot-Org/KittyBotGo/db/.gen/kittybot-go/public/model"
	"github.com/KittyBot-Org/KittyBotGo/dbot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/json"
	"github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/snowflake/v2"
	"golang.org/x/text/message"
)

var Report = handler.Command{
	Create: discord.MessageCommandCreate{
		Name: "report",
	},
	CommandHandlers: map[string]handler.CommandHandler{
		"": reportHandler,
	},
	ComponentHandler: map[string]dbot.ComponentHandler{
		"action":  reportActionHandler,
		"confirm": reportConfirmHandler,
		"delete":  reportDeleteHandler,
	},
	ModalHandler: map[string]dbot.ModalHandler{
		"action-confirm": reportActionConfirmHandler,
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

func CreateReport(b *dbot.Bot, settings model.GuildSetting, reportID int32, content string, embed discord.Embed) error {
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
					CustomID: fmt.Sprintf("cmd:report:confirm:%d", reportID),
				},
				discord.ButtonComponent{
					Style:    discord.ButtonStyleDanger,
					Label:    "Delete",
					CustomID: fmt.Sprintf("cmd:report:delete:%d", reportID),
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

	report, err := b.DB.Reports().Get(reportID)
	if err != nil {
		return err
	}

	reportCount, err := b.DB.Reports().GetCount(snowflake.MustParse(report.UserID), snowflake.MustParse(report.GuildID))
	if err != nil {
		b.Logger.Errorf("Failed to get report count: %s", err)
		return e.CreateMessage(discord.MessageCreate{
			Content: "Failed to get report count, please reach out to a bot developer.",
			Flags:   discord.MessageFlagEphemeral,
		})
	}

	if err = b.DB.Reports().Confirm(reportID); err != nil {
		b.Logger.Errorf("Failed to confirm report: %s", err)
		return e.CreateMessage(discord.MessageCreate{
			Content: "Failed to confirm report, please reach out to a bot developer.",
			Flags:   discord.MessageFlagEphemeral,
		})
	}

	var selectMenuOptions []discord.SelectMenuOption

	if reportCount > 0 {
		selectMenuOptions = append(selectMenuOptions, discord.SelectMenuOption{
			Label: "Show Previous Reports",
			Value: "show-reports",
			Emoji: &discord.ComponentEmoji{
				Name: "📜",
			},
		})
	}

	if report.MessageID != "0" && report.ChannelID != "0" {
		selectMenuOptions = append(selectMenuOptions, discord.SelectMenuOption{
			Label: "Delete Message",
			Value: "delete-message",
			Emoji: &discord.ComponentEmoji{
				Name: "🗑",
			},
		})
	}

	selectMenuOptions = append(selectMenuOptions, []discord.SelectMenuOption{
		{
			Label: "Delete Report",
			Value: "delete-report",
			Emoji: &discord.ComponentEmoji{
				Name: "🗑",
			},
		},
		{
			Label: "Timeout User",
			Value: "timeout",
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
	}...)

	return e.UpdateMessage(discord.MessageUpdate{
		Components: &[]discord.ContainerComponent{
			discord.ActionRowComponent{
				discord.SelectMenuComponent{
					CustomID:    fmt.Sprintf("cmd:report:action:%d", reportID),
					Placeholder: "Select an action",
					MinValues:   json.NewPtr(1),
					MaxValues:   1,
					Options:     selectMenuOptions,
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

	reason := fmt.Sprintf("AutoMod action by: %s\nCaused by report #%d", e.User().Tag(), reportID)

	var content string
	value := e.SelectMenuInteractionData().Values[0]
	switch value {
	case "delete-message":
		if err = b.Client.Rest().DeleteMessage(snowflake.MustParse(report.ChannelID), snowflake.MustParse(report.MessageID), rest.WithReason(reason)); err != nil {
			b.Logger.Errorf("Failed to delete message: %s", err)
			content = "Failed to delete message, please reach out to a bot developer."
		} else {
			content = "Message deleted."
		}

	case "delete-report":
		if err = b.DB.Reports().Delete(reportID); err != nil {
			b.Logger.Errorf("Failed to delete report: %s", err)
			content = "Failed to delete report, please reach out to a bot developer."
		} else {
			content = "Report deleted."
		}

		if err = e.DeferUpdateMessage(); err != nil {
			b.Logger.Errorf("Failed to defer update message: %s", err)
			return err
		}

		if err = e.Client().Rest().DeleteInteractionResponse(e.ApplicationID(), e.Token()); err != nil {
			b.Logger.Errorf("Failed to delete interaction response: %s", err)
			content = "Failed to delete interaction response, please reach out to a bot developer."
		}

		_, err = b.Client.Rest().CreateFollowupMessage(e.ApplicationID(), e.Token(), discord.MessageCreate{
			Content: content,
			Flags:   discord.MessageFlagEphemeral,
		})
		return err

	case "show-reports":
		reports, err := b.DB.Reports().GetAll(snowflake.MustParse(report.UserID), snowflake.MustParse(report.GuildID))
		if err == sql.ErrNoRows {
			content = "No reports found."
		} else if err != nil {
			b.Logger.Errorf("Failed to get reports: %s", err)
			content = "Failed to get reports, please reach out to a bot developer."
		} else {
			user, err := b.Client.Rest().GetUser(snowflake.MustParse(report.UserID))
			if err != nil {
				user = &discord.User{
					Username:      "Unknown",
					Discriminator: "xxxx",
				}
			}
			content = formatReports(reports, *user)
		}

	case "timeout":
		return e.CreateModal(discord.ModalCreate{
			CustomID: fmt.Sprintf("cmd:report:action-confirm:timeout:%s", report.UserID),
			Title:    "Timeout User",
			Components: []discord.ContainerComponent{
				discord.ActionRowComponent{
					discord.TextInputComponent{
						CustomID:    "duration",
						Style:       discord.TextInputStyleShort,
						Label:       "Timeout Duration",
						MinLength:   json.NewPtr(2),
						Required:    true,
						Placeholder: "time units: s, m, h example: 1h3s",
						Value:       "1h",
					},
				},
				discord.ActionRowComponent{
					discord.TextInputComponent{
						CustomID:    "reason",
						Style:       discord.TextInputStyleParagraph,
						Label:       "Timeout Reason",
						Required:    true,
						Placeholder: "The reason for the timeout",
						Value:       reason,
					},
				},
			},
		})

	case "kick":
		return e.CreateModal(discord.ModalCreate{
			CustomID: fmt.Sprintf("cmd:report:action-confirm:kick:%s", report.UserID),
			Title:    "Kick User",
			Components: []discord.ContainerComponent{
				discord.ActionRowComponent{
					discord.TextInputComponent{
						CustomID:    "reason",
						Style:       discord.TextInputStyleParagraph,
						Label:       "Kick Reason",
						Required:    true,
						Placeholder: "The reason for the kick",
						Value:       reason,
					},
				},
			},
		})

	case "ban":
		return e.CreateModal(discord.ModalCreate{
			CustomID: fmt.Sprintf("cmd:report:action-confirm:ban:%s", report.UserID),
			Title:    "Ban User",
			Components: []discord.ContainerComponent{
				discord.ActionRowComponent{
					discord.TextInputComponent{
						CustomID:    "del-days",
						Style:       discord.TextInputStyleShort,
						Label:       "Message Delete Days",
						MinLength:   json.NewPtr(1),
						MaxLength:   1,
						Required:    true,
						Placeholder: "0-7",
						Value:       "0",
					},
				},
				discord.ActionRowComponent{
					discord.TextInputComponent{
						CustomID:    "reason",
						Style:       discord.TextInputStyleParagraph,
						Label:       "Ban Reason",
						Required:    true,
						Placeholder: "The reason for the ban",
						Value:       reason,
					},
				},
			},
		})

	default:
		b.Logger.Errorf("Unknown report action: %s", value)
		content = "Unknown action."
	}
	return e.CreateMessage(discord.MessageCreate{
		Content: content,
		Flags:   discord.MessageFlagEphemeral,
	})
}

func reportActionConfirmHandler(b *dbot.Bot, args []string, p *message.Printer, e *events.ModalSubmitInteractionCreate) error {
	userID := snowflake.MustParse(args[1])
	reason := e.Data.Text("reason")

	var content string
	switch args[0] {
	case "timeout":
		until := time.Now()
		duration, err := time.ParseDuration(e.Data.Text("duration"))
		if err != nil {
			content = "Invalid duration. Please use a valid duration."
		} else {
			until = until.Add(duration)
			if _, err = b.Client.Rest().UpdateMember(*e.GuildID(), userID, discord.MemberUpdate{
				CommunicationDisabledUntil: json.NewOptional(until),
			}, rest.WithReason(reason)); err != nil {
				b.Logger.Errorf("Failed to update member: %s", err)
				content = "Failed to timeout user, please reach out to a bot developer."
			} else {
				content = fmt.Sprintf("Timed out user until %s.", discord.TimestampStyleShortDateTime.FormatTime(until))
			}
		}

	case "kick":
		if err := b.Client.Rest().RemoveMember(*e.GuildID(), userID, rest.WithReason(reason)); err != nil {
			b.Logger.Errorf("Failed to kick user: %s", err)
			content = "Failed to kick user, please reach out to a bot developer."
		} else {
			content = "User kicked."
		}

	case "ban":
		delDays, err := strconv.Atoi(e.Data.Text("del-days"))
		if err != nil || delDays < 0 || delDays > 7 {
			content = "Invalid message deletion days. Make sure it's a number between 0 and 7."
		} else {
			if err = b.Client.Rest().AddBan(*e.GuildID(), userID, 0, rest.WithReason(reason)); err != nil {
				b.Logger.Errorf("Failed to ban user: %s", err)
				content = "Failed to ban user, please reach out to a bot developer."
			} else {
				content = fmt.Sprintf("Banned %s. And deleted messages of the last %d days.", discord.UserMention(userID), delDays)
			}
		}
	}

	return e.CreateMessage(discord.MessageCreate{
		Content: content,
		Flags:   discord.MessageFlagEphemeral,
	})
}
