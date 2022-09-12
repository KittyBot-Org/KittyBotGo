package commands

import (
	"github.com/KittyBot-Org/KittyBotGo/dbot"
	"github.com/KittyBot-Org/KittyBotGo/dbot/responses"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"golang.org/x/text/message"
)

var Next = handler.Command{
	Create: discord.SlashCommandCreate{
		Name:        "next",
		Description: "Stops the song and starts the next one.",
	},
	Checks: dbot.HasMusicPlayer.And(dbot.IsMemberConnectedToVoiceChannel).And(dbot.HasQueueItems),
	CommandHandler: map[string]handler.CommandHandler{
		"": nextHandler,
	},
}

func nextHandler(b *dbot.Bot, p *message.Printer, e *events.ApplicationCommandInteractionCreate) error {
	player := b.MusicPlayers.Get(*e.GuildID())
	nextTrack := player.Queue.Pop()

	if nextTrack == nil {
		return e.CreateMessage(responses.CreateErrorf(p, "modules.music.commands.next.no.track"))
	}

	if err := player.Play(nextTrack); err != nil {
		return e.CreateMessage(responses.CreateErrorf(p, "modules.music.commands.next.error"))
	}
	return e.CreateMessage(responses.CreateSuccessComponentsf(p, "modules.music.commands.next.success", []any{formatTrack(nextTrack), nextTrack.Info().Length}, getMusicControllerComponents(nextTrack)))
}
