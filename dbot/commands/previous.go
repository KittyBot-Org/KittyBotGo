package commands

import (
	"github.com/KittyBot-Org/KittyBotGo/dbot"
	"github.com/KittyBot-Org/KittyBotGo/dbot/responses"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"golang.org/x/text/message"
)

var Previous = dbot.Command{
	Create: discord.SlashCommandCreate{
		CommandName: "previous",
		Description: "Stops the song and starts the previous one.",
	},
	Checks: dbot.HasMusicPlayer.And(dbot.IsMemberConnectedToVoiceChannel).And(dbot.HasHistoryItems),
	CommandHandler: map[string]dbot.CommandHandler{
		"": previousHandler,
	},
}

func previousHandler(b *dbot.Bot, p *message.Printer, e *events.ApplicationCommandInteractionCreate) error {
	player := b.MusicPlayers.Get(*e.GuildID())
	previousTrack := player.History.Last()

	if previousTrack == nil {
		return e.CreateMessage(responses.CreateErrorf(p, "modules.music.commands.previous.no.track"))
	}

	if err := player.Play(previousTrack); err != nil {
		return e.CreateMessage(responses.CreateErrorf(p, "modules.music.commands.previous.error"))
	}
	return e.CreateMessage(responses.CreateSuccessComponentsf(p, "modules.music.commands.previous.success", []any{formatTrack(previousTrack), previousTrack.Info().Length}, getMusicControllerComponents(previousTrack)))
}
