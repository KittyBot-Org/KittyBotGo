package commands

import (
	"database/sql"
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/handler"
	"github.com/disgoorg/json"

	"github.com/KittyBot-Org/KittyBotGo/db/.gen/kittybot-go/public/model"
	"github.com/KittyBot-Org/KittyBotGo/dbot"
)

func Reports(b *dbot.Bot) handler.Command {
	return handler.Command{
		Create: discord.SlashCommandCreate{
			Name:        "reports",
			Description: "View/Delete reports of a user.",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionSubCommand{
					Name:        "view",
					Description: "View a report of a user.",
					Options: []discord.ApplicationCommandOption{
						discord.ApplicationCommandOptionUser{
							Name:        "user",
							Description: "The user to view a report of.",
							Required:    true,
						},
						discord.ApplicationCommandOptionInt{
							Name:         "report",
							Description:  "The user to view a report of.",
							Required:     true,
							Autocomplete: true,
						},
					},
				},
				discord.ApplicationCommandOptionSubCommand{
					Name:        "view-all",
					Description: "View all reports of a user.",
					Options: []discord.ApplicationCommandOption{
						discord.ApplicationCommandOptionUser{
							Name:        "user",
							Description: "The user to view reports of.",
							Required:    true,
						},
					},
				},
				discord.ApplicationCommandOptionSubCommand{
					Name:        "delete",
					Description: "Delete a reports of a user.",
					Options: []discord.ApplicationCommandOption{
						discord.ApplicationCommandOptionUser{
							Name:        "user",
							Description: "The user to delete a report of.",
							Required:    true,
						},
						discord.ApplicationCommandOptionInt{
							Name:         "report",
							Description:  "The user to view reports of.",
							Required:     true,
							Autocomplete: true,
						},
					},
				},
				discord.ApplicationCommandOptionSubCommand{
					Name:        "delete-all",
					Description: "Deletes all reports of a user.",
					Options: []discord.ApplicationCommandOption{
						discord.ApplicationCommandOptionUser{
							Name:        "user",
							Description: "The user to view reports of.",
							Required:    true,
						},
					},
				},
			},
			DefaultMemberPermissions: json.NewNullablePtr(discord.PermissionKickMembers | discord.PermissionBanMembers | discord.PermissionModerateMembers),
		},
		CommandHandlers: map[string]handler.CommandHandler{
			"view":       reportsViewHandler(b),
			"view-all":   reportsViewAllHandler(b),
			"delete":     reportsDeleteHandler(b),
			"delete-all": reportsDeleteAllHandler(b),
		},
		AutocompleteHandlers: map[string]handler.AutocompleteHandler{
			"view":   reportAutocompleteReportHandler(b),
			"delete": reportAutocompleteReportHandler(b),
		},
	}
}

func reportsViewHandler(b *dbot.Bot) handler.CommandHandler {
	return func(e *events.ApplicationCommandInteractionCreate) error {
		data := e.SlashCommandInteractionData()
		reportID := int32(data.Int("report"))

		report, err := b.DB.Reports().Get(reportID)
		if err == sql.ErrNoRows {
			return e.CreateMessage(discord.MessageCreate{
				Content: "Report not found.",
				Flags:   discord.MessageFlagEphemeral,
			})
		} else if err != nil {
			b.Logger.Errorf("Error getting report: %s", err)
			return e.CreateMessage(discord.MessageCreate{
				Content: "Failed to get report, please reach out to a bot developer.",
				Flags:   discord.MessageFlagEphemeral,
			})
		}

		return e.CreateMessage(discord.MessageCreate{
			Content: fmt.Sprintf("**Report #%d from %s:**\n%s", report.ID, discord.TimestampStyleShortDateTime.FormatTime(report.CreatedAt), report.Description),
			Flags:   discord.MessageFlagEphemeral,
		})
	}
}

func formatReports(reports []model.Reports, user discord.User) string {
	content := fmt.Sprintf("**%d Reports for %s:**\n", len(reports), user.Tag())
	for i, report := range reports {
		newLine := fmt.Sprintf("%d. %s, %s\n", i+1, discord.TimestampStyleShortDateTime.FormatTime(report.CreatedAt), trimString(report.Description, 20))
		if len(content+newLine) > 2000 {
			break
		}
		content += newLine
	}
	return content
}

func reportsViewAllHandler(b *dbot.Bot) handler.CommandHandler {
	return func(e *events.ApplicationCommandInteractionCreate) error {
		data := e.SlashCommandInteractionData()
		userID := data.Snowflake("user")

		reports, err := b.DB.Reports().GetAll(userID, *e.GuildID())
		if err == sql.ErrNoRows {
			return e.CreateMessage(discord.MessageCreate{
				Content: "No reports found.",
				Flags:   discord.MessageFlagEphemeral,
			})
		} else if err != nil {
			b.Logger.Errorf("Error getting reports: %s", err)
			return e.CreateMessage(discord.MessageCreate{
				Content: "Failed to get reports, please reach out to a bot developer.",
				Flags:   discord.MessageFlagEphemeral,
			})
		}

		content := formatReports(reports, data.User("user"))

		return e.CreateMessage(discord.MessageCreate{
			Content: content,
			Flags:   discord.MessageFlagEphemeral,
		})
	}
}

func trimString(s string, max int) string {
	if len(s) > max {
		return s[:max-1] + "…"
	}
	return s
}

func reportsDeleteHandler(b *dbot.Bot) handler.CommandHandler {
	return func(e *events.ApplicationCommandInteractionCreate) error {
		data := e.SlashCommandInteractionData()
		reportID := int32(data.Int("report"))

		err := b.DB.Reports().Delete(reportID)
		if err != nil {
			b.Logger.Errorf("Error deleting report: %s", err)
			return e.CreateMessage(discord.MessageCreate{
				Content: "Failed to delete report, please reach out to a bot developer.",
				Flags:   discord.MessageFlagEphemeral,
			})
		}

		return e.CreateMessage(discord.MessageCreate{
			Content: "Report deleted.",
			Flags:   discord.MessageFlagEphemeral,
		})
	}
}

func reportsDeleteAllHandler(b *dbot.Bot) handler.CommandHandler {
	return func(e *events.ApplicationCommandInteractionCreate) error {
		data := e.SlashCommandInteractionData()
		userID := data.Snowflake("user")

		err := b.DB.Reports().DeleteAll(userID, *e.GuildID())
		if err != nil {
			b.Logger.Errorf("Error deleting reports: %s", err)
			return e.CreateMessage(discord.MessageCreate{
				Content: "Failed to delete reports, please reach out to a bot developer.",
				Flags:   discord.MessageFlagEphemeral,
			})
		}

		return e.CreateMessage(discord.MessageCreate{
			Content: "All reports deleted.",
			Flags:   discord.MessageFlagEphemeral,
		})
	}
}

func reportAutocompleteReportHandler(b *dbot.Bot) handler.AutocompleteHandler {
	return func(e *events.AutocompleteInteractionCreate) error {
		data := e.Data
		userID := data.Snowflake("user")

		reports, err := b.DB.Reports().GetAll(userID, *e.GuildID())
		if err != nil {
			b.Logger.Errorf("Error getting reports: %s", err)
			return e.Result(nil)
		}

		var choices []discord.AutocompleteChoice
		for _, report := range reports {
			choices = append(choices, discord.AutocompleteChoiceInt{
				Name:  fmt.Sprintf("#%d", report.ID),
				Value: int(report.ID),
			})
		}
		return e.Result(choices)
	}
}
