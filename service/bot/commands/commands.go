package commands

import (
	"github.com/disgoorg/disgo/discord"

	"github.com/KittyBot-Org/KittyBotGo/service/bot"
)

var Commands = []discord.ApplicationCommandCreate{
	ping,
	play,
}

func New(b *bot.Bot) *Bot {
	return &Bot{
		Bot: b,
	}
}

type Bot struct {
	*bot.Bot
}
