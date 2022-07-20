package commands

import (
	"github.com/KittyBot-Org/KittyBotGo/dbot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"golang.org/x/text/message"
)

var ClearQueue = dbot.Command{
	Create: discord.SlashCommandCreate{
		CommandName: "clear-queue",
		Description: "Removes all songs from your queue.",
	},
	Checks: dbot.HasMusicPlayer.And(dbot.IsMemberConnectedToVoiceChannel).And(dbot.HasQueueItems),
	CommandHandler: map[string]dbot.CommandHandler{
		"": clearQueueHandler,
	},
}

func clearQueueHandler(b *dbot.Bot, p *message.Printer, e *events.ApplicationCommandInteractionCreate) error {
	b.MusicPlayers.Get(*e.GuildID()).Queue.Clear()
	return e.CreateMessage(discord.MessageCreate{Content: p.Sprintf("modules.music.commands.clear.cleared")})
}
