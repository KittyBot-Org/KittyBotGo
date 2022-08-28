package commands

import (
	"github.com/KittyBot-Org/KittyBotGo/dbot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/json"
	"github.com/disgoorg/snowflake/v2"
	"golang.org/x/text/message"
)

var Settings = dbot.Command{
	Create: discord.SlashCommandCreate{
		Name:        "settings",
		Description: "View and edit settings",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionSubCommand{
				Name:        "view",
				Description: "View all settings",
			},
			discord.ApplicationCommandOptionSubCommandGroup{
				Name:        "moderation",
				Description: "Moderation settings",
				Options: []discord.ApplicationCommandOptionSubCommand{
					{
						Name:        "disable",
						Description: "Disables moderation",
					},
					{
						Name:        "log-channel",
						Description: "Set the channel to log moderation actions",
						Options: []discord.ApplicationCommandOption{
							discord.ApplicationCommandOptionChannel{
								Name:        "channel",
								Description: "The channel to log moderation actions to",
								Required:    true,
							},
						},
					},
				},
			},
		},
		DefaultMemberPermissions: discord.PermissionManageServer,
	},
	CommandHandler: map[string]dbot.CommandHandler{
		"moderation/disable":     settingsModerationDisableHandler,
		"moderation/log-channel": settingsModerationLogChannelHandler,
	},
}

func settingsModerationDisableHandler(b *dbot.Bot, p *message.Printer, e *events.ApplicationCommandInteractionCreate) error {
	if err := b.DB.GuildSettings().UpdateModeration(*e.GuildID(), 0, ""); err != nil {
		b.Logger.Errorf("Error updating guild settings: %s", err)
		return e.CreateMessage(discord.MessageCreate{
			Content: "Error setting settings, please reach out to a bot developer.",
			Flags:   discord.MessageFlagEphemeral,
		})
	}

	return e.CreateMessage(discord.MessageCreate{
		Content: "Moderation disabled",
		Flags:   discord.MessageFlagEphemeral,
	})
}

func settingsModerationLogChannelHandler(b *dbot.Bot, p *message.Printer, e *events.ApplicationCommandInteractionCreate) error {
	data := e.SlashCommandInteractionData()
	settings, err := b.DB.GuildSettings().Get(*e.GuildID())
	if err != nil {
		b.Logger.Errorf("Error getting guild settings: %s", err)
		return e.CreateMessage(discord.MessageCreate{
			Content: "Error getting settings, please reach out to a bot developer.",
			Flags:   discord.MessageFlagEphemeral,
		})
	}

	if settings.ModerationLogWebhookID == "0" || settings.ModerationLogWebhookToken == "" {
		incomingWebhook, err := b.Client.Rest().CreateWebhook(data.Snowflake("channel"), discord.WebhookCreate{
			Name: "Automod",
		})
		if err != nil {
			b.Logger.Errorf("Error creating webhook: %s", err)
			return e.CreateMessage(discord.MessageCreate{
				Content: "Error creating webhook, please reach out to a bot developer.",
				Flags:   discord.MessageFlagEphemeral,
			})
		}
		settings.ModerationLogWebhookID = incomingWebhook.ID().String()
		settings.ModerationLogWebhookToken = incomingWebhook.Token

		if err = b.DB.GuildSettings().UpdateModeration(*e.GuildID(), incomingWebhook.ID(), incomingWebhook.Token); err != nil {
			b.Logger.Errorf("Error updating guild settings: %s", err)
			return e.CreateMessage(discord.MessageCreate{
				Content: "Error updating guild settings, please reach out to a bot developer.",
				Flags:   discord.MessageFlagEphemeral,
			})
		}
	} else {
		if _, err = b.Client.Rest().UpdateWebhook(snowflake.MustParse(settings.ModerationLogWebhookID), discord.WebhookUpdate{
			ChannelID: json.NewPtr(data.Snowflake("channel")),
		}); err != nil {
			b.Logger.Errorf("Error updating webhook: %s", err)
			return e.CreateMessage(discord.MessageCreate{
				Content: "Error updating existing webhook, please reach out to a bot developer.",
				Flags:   discord.MessageFlagEphemeral,
			})
		}
	}
	client := b.ReportLogWebhookMap.Get(snowflake.MustParse(settings.ModerationLogWebhookID), settings.ModerationLogWebhookToken)
	if _, err = client.CreateMessage(discord.WebhookMessageCreate{
		Content: "Moderation log channel successfully set",
	}); err != nil {
		b.Logger.Errorf("Error creating message: %s", err)
	}

	return e.CreateMessage(discord.MessageCreate{
		Content: "Moderation log channel updated",
		Flags:   discord.MessageFlagEphemeral,
	})
}
