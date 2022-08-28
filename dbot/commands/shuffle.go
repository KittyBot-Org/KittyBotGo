package commands

import (
	"github.com/KittyBot-Org/KittyBotGo/dbot"
	"github.com/KittyBot-Org/KittyBotGo/dbot/responses"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"golang.org/x/text/message"
)

var Shuffle = dbot.Command{
	Create: discord.SlashCommandCreate{
		Name:        "shuffle",
		Description: "Shuffles the queue of songs.",
	},
	Checks: dbot.HasMusicPlayer.And(dbot.IsMemberConnectedToVoiceChannel).And(dbot.HasQueueItems),
	CommandHandler: map[string]dbot.CommandHandler{
		"": shuffleHandler,
	},
}

func shuffleHandler(b *dbot.Bot, p *message.Printer, e *events.ApplicationCommandInteractionCreate) error {
	queue := b.MusicPlayers.Get(*e.GuildID()).Queue

	if queue.Len() == 0 {
		return e.CreateMessage(responses.CreateErrorf(p, "modules.music.commands.shuffle.no.track"))
	}
	return e.CreateMessage(responses.CreateSuccessf(p, "modules.music.commands.shuffle.success"))
}
