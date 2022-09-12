package commands

import (
	"github.com/KittyBot-Org/KittyBotGo/dbot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/handler"
)

func ClearQueue(b *dbot.Bot) handler.Command {
	return handler.Command{
		Create: discord.SlashCommandCreate{
			Name:        "clear-queue",
			Description: "Removes all songs from your queue.",
		},
		Check: dbot.HasMusicPlayer(b).And(dbot.IsMemberConnectedToVoiceChannel(b)).And(dbot.HasQueueItems(b)),
		CommandHandlers: map[string]handler.CommandHandler{
			"": clearQueueHandler(b),
		},
	}
}

func clearQueueHandler(b *dbot.Bot) handler.CommandHandler {
	return func(e *events.ApplicationCommandInteractionCreate) error {
		b.MusicPlayers.Get(*e.GuildID()).Queue.Clear()
		return e.CreateMessage(discord.MessageCreate{Content: "modules.music.commands.clear.cleared"})
	}
}
