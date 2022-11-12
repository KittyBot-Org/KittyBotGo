package commands

import (
	"github.com/KittyBot-Org/KittyBotGo/dbot"
	"github.com/KittyBot-Org/KittyBotGo/dbot/responses"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/handler"
)

func Shuffle(b *dbot.Bot) handler.Command {
	return handler.Command{
		Create: discord.SlashCommandCreate{
			Name:        "shuffle",
			Description: "Shuffles the queue of songs.",
		},
		Check: dbot.HasMusicPlayer(b).And(dbot.IsMemberConnectedToVoiceChannel(b)).And(dbot.HasQueueItems(b)),
		CommandHandlers: map[string]handler.CommandHandler{
			"": shuffleHandler(b),
		},
	}
}

func shuffleHandler(b *dbot.Bot) handler.CommandHandler {
	return func(e *events.ApplicationCommandInteractionCreate) error {
		queue := b.MusicPlayers.Get(*e.GuildID()).Queue

		if queue.Len() == 0 {
			return e.CreateMessage(responses.CreateErrorf("No songs in queue to shuffle."))
		}
		return e.CreateMessage(responses.CreateSuccessf("🔀 Shuffled the queue."))
	}
}
