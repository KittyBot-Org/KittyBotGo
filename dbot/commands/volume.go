package commands

import (
	"github.com/KittyBot-Org/KittyBotGo/dbot"
	"github.com/KittyBot-Org/KittyBotGo/dbot/responses"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/json"
	"github.com/disgoorg/handler"
)

func Volume(bot *dbot.Bot) handler.Command {
	return handler.Command{
		Create: discord.SlashCommandCreate{
			Name:        "volume",
			Description: "Changes the volume of the music player.",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionInt{
					Name:        "volume",
					Description: "the desired volume",
					Required:    true,
					MinValue:    json.NewPtr(0),
					MaxValue:    json.NewPtr(100),
				},
			},
		},
		Check: dbot.HasMusicPlayer(bot).And(dbot.IsMemberConnectedToVoiceChannel(bot)),
		CommandHandlers: map[string]handler.CommandHandler{
			"": volumeHandler(bot),
		},
	}
}

func volumeHandler(bot *dbot.Bot) handler.CommandHandler {
	return func(e *events.ApplicationCommandInteractionCreate) error {
		player := bot.MusicPlayers.Get(*e.GuildID())
		volume := e.SlashCommandInteractionData().Int("volume")
		if err := player.SetVolume(volume); err != nil {
			return e.CreateMessage(responses.CreateErrorf("modules.music.commands.volume.error"))
		}
		return e.CreateMessage(responses.CreateSuccessf("modules.music.commands.volume.success", volume))
	}
}
