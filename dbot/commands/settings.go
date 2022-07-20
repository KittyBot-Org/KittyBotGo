package commands

import (
	"github.com/KittyBot-Org/KittyBotGo/dbot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"golang.org/x/text/message"
)

var Settings = dbot.Command{
	Create: discord.SlashCommandCreate{
		CommandName: "settings",
		Description: "View and edit settings",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionSubCommandGroup{
				GroupName:   "moderation",
				Description: "Moderation settings",
				Options: []discord.ApplicationCommandOptionSubCommand{
					{
						CommandName: "log-channel",
						Description: "Set the channel to log moderation actions",
						Options: []discord.ApplicationCommandOption{
							discord.ApplicationCommandOptionChannel{
								OptionName:  "channel",
								Description: "The channel to log moderation actions to",
							},
						},
					},
				},
			},
		},
		DefaultMemberPermissions: discord.PermissionManageServer,
	},
	CommandHandler: map[string]dbot.CommandHandler{
		"moderation/log-channel": settingsModerationLogChannelHandler,
	},
}

func settingsModerationLogChannelHandler(b *dbot.Bot, p *message.Printer, e *events.ApplicationCommandInteractionCreate) error {
	return nil
}
