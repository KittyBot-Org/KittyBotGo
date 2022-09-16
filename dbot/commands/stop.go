package commands

import (
	"context"

	"github.com/KittyBot-Org/KittyBotGo/dbot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/handler"
)

func Stop(b *dbot.Bot) handler.Command {
	return handler.Command{
		Create: discord.SlashCommandCreate{
			Name:        "stop",
			Description: "Stops the playing music.",
		},
		Check: dbot.HasMusicPlayer(b).And(dbot.IsMemberConnectedToVoiceChannel(b)),
		CommandHandlers: map[string]handler.CommandHandler{
			"": stopHandler(b),
		},
	}
}

func stopHandler(b *dbot.Bot) handler.CommandHandler {
	return func(e *events.ApplicationCommandInteractionCreate) error {
		player := b.MusicPlayers.Get(*e.GuildID())
		if err := player.Destroy(); err != nil {
			b.Logger.Error("Failed to destroy player: ", err)
			err = e.CreateMessage(discord.MessageCreate{
				Content: "Failed to stop the music player. Please try again.",
				Flags:   discord.MessageFlagEphemeral,
			})
			if err != nil {
				b.Logger.Error("Failed to send message: ", err)
			}
			return err
		}
		if err := b.Client.Disconnect(context.TODO(), *e.GuildID()); err != nil {
			b.Logger.Error("Failed to disconnect dbot: ", err)
			err = e.CreateMessage(discord.MessageCreate{
				Content: "Failed to disconnect from voice channel. Please try again.",
				Flags:   discord.MessageFlagEphemeral,
			})
			if err != nil {
				b.Logger.Error("Failed to send message: ", err)
			}
			return err
		}
		b.MusicPlayers.Delete(*e.GuildID())
		return e.CreateMessage(discord.MessageCreate{Content: "Stopped the player."})
	}
}
