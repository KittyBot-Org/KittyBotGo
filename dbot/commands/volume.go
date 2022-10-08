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
					Description: "The new volume",
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
			return e.CreateMessage(responses.CreateErrorf("Failed to set the volume. Please try again."))
		}
		return e.CreateMessage(responses.CreateSuccessf("🔊 Volume set to `%d`.", volume))
	}
}
