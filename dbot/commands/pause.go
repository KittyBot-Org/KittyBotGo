package commands

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/handler"

	"github.com/KittyBot-Org/KittyBotGo/dbot"
	"github.com/KittyBot-Org/KittyBotGo/dbot/responses"
)

func Pause(b *dbot.Bot) handler.Command {
	return handler.Command{
		Create: discord.SlashCommandCreate{
			Name:        "pause",
			Description: "Pauses or resumes the player.",
		},
		Check: dbot.HasMusicPlayer(b).And(dbot.IsMemberConnectedToVoiceChannel(b)),
		CommandHandlers: map[string]handler.CommandHandler{
			"": pauseHandler(b),
		},
	}
}

func pauseHandler(b *dbot.Bot) handler.CommandHandler {
	return func(e *events.ApplicationCommandInteractionCreate) error {
		player := b.MusicPlayers.Get(*e.GuildID())
		pause := !player.Paused()
		if err := player.Pause(pause); err != nil {
			var msg string
			if pause {
				msg = "Failed to pause the player. Please try again."
			} else {
				msg = "Failed to resume the player. Please try again."
			}
			return e.CreateMessage(responses.CreateSuccessf(msg))
		}
		var msg string
		if pause {
			msg = "⏯ Paused the player."
		} else {
			msg = "⏯ Resumed the player."
		}
		return e.CreateMessage(responses.CreateSuccessComponentsf(msg, nil, getMusicControllerComponents(player.PlayingTrack())))
	}
}
