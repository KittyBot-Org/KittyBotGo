package commands

import (
	"github.com/KittyBot-Org/KittyBotGo/dbot"
	"github.com/KittyBot-Org/KittyBotGo/dbot/responses"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/handler"
)

func Previous(b *dbot.Bot) handler.Command {
	return handler.Command{
		Create: discord.SlashCommandCreate{
			Name:        "previous",
			Description: "Stops the song and starts the previous one.",
		},
		Check: dbot.HasMusicPlayer(b).And(dbot.IsMemberConnectedToVoiceChannel(b)).And(dbot.HasHistoryItems(b)),
		CommandHandlers: map[string]handler.CommandHandler{
			"": previousHandler(b),
		},
	}
}

func previousHandler(b *dbot.Bot) handler.CommandHandler {
	return func(e *events.ApplicationCommandInteractionCreate) error {
		player := b.MusicPlayers.Get(*e.GuildID())
		previousTrack := player.History.Last()

		if previousTrack == nil {
			return e.CreateMessage(responses.CreateErrorf("No track found in history."))
		}

		if err := player.Play(previousTrack); err != nil {
			return e.CreateMessage(responses.CreateErrorf("Failed to play previous song. Please try again."))
		}
		return e.CreateMessage(responses.CreateSuccessComponentsf("⏮ Skipped to previous song.\nNow playing: %s - %s", []any{formatTrack(previousTrack), previousTrack.Info().Length}, getMusicControllerComponents(previousTrack)))
	}
}
