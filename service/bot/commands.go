package bot

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
)

var commands = []discord.ApplicationCommandCreate{
	discord.SlashCommandCreate{
		Name:        "ping",
		Description: "Ping the bot",
	},
}

func (b *Bot) OnPing(e *handler.CommandEvent) error {
	return e.CreateMessage(discord.MessageCreate{
		Content: "Pong!",
	})
}
