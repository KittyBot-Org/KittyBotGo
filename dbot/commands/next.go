package commands

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/handler"

	"github.com/KittyBot-Org/KittyBotGo/dbot"
	"github.com/KittyBot-Org/KittyBotGo/dbot/responses"
)

func Next(b *dbot.Bot) handler.Command {
	return handler.Command{
		Create: discord.SlashCommandCreate{
			Name:        "next",
			Description: "Stops the track and starts the next one.",
		},
		Check: dbot.HasMusicPlayer(b).And(dbot.IsMemberConnectedToVoiceChannel(b)).And(dbot.HasQueueItems(b)),
		CommandHandlers: map[string]handler.CommandHandler{
			"": nextHandler(b),
		},
	}
}

func nextHandler(b *dbot.Bot) handler.CommandHandler {
	return func(e *events.ApplicationCommandInteractionCreate) error {
		player := b.MusicPlayers.Get(*e.GuildID())
		nextTrack := player.Queue.Pop()

		if nextTrack == nil {
			return e.CreateMessage(responses.CreateErrorf("No next track found in queue."))
		}

		if err := player.Play(nextTrack); err != nil {
			return e.CreateMessage(responses.CreateErrorf("Failed to play next track. Please try again."))
		}
		return e.CreateMessage(responses.CreateSuccessComponentsf("⏭ Skipped to next track.\nNow playing: %s - %s", []any{formatTrack(nextTrack), nextTrack.Info().Length}, getMusicControllerComponents(nextTrack)))
	}
}
