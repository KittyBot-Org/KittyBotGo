package commands

import (
	"context"

	"github.com/KittyBot-Org/KittyBotGo/dbot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"golang.org/x/text/message"
)

var Stop = handler.Command{
	Create: discord.SlashCommandCreate{
		Name:        "stop",
		Description: "Stops the playing music.",
	},
	Checks: dbot.HasMusicPlayer.And(dbot.IsMemberConnectedToVoiceChannel),
	CommandHandler: map[string]handler.CommandHandler{
		"": stopHandler,
	},
}

func stopHandler(b *dbot.Bot, p *message.Printer, e *events.ApplicationCommandInteractionCreate) error {
	player := b.MusicPlayers.Get(*e.GuildID())
	if err := player.Destroy(); err != nil {
		b.Logger.Error("Failed to destroy player: ", err)
		err = e.CreateMessage(discord.MessageCreate{Content: p.Sprintf("modules.music.commands.stop.error")})
		if err != nil {
			b.Logger.Error("Failed to send message: ", err)
		}
		return err
	}
	if err := b.Client.Disconnect(context.TODO(), *e.GuildID()); err != nil {
		b.Logger.Error("Failed to disconnect dbot: ", err)
		err = e.CreateMessage(discord.MessageCreate{Content: p.Sprintf("modules.music.commands.stop.disconnect.error")})
		if err != nil {
			b.Logger.Error("Failed to send message: ", err)
		}
		return err
	}
	b.MusicPlayers.Delete(*e.GuildID())
	return e.CreateMessage(discord.MessageCreate{Content: p.Sprintf("modules.music.commands.stop.stopped")})
}
