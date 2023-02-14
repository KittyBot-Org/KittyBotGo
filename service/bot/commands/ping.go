package commands

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"

	"github.com/KittyBot-Org/KittyBotGo/service/bot/res"
)

var ping = discord.SlashCommandCreate{
	Name:        "ping",
	Description: "Ping the bot",
}

func (c *Cmds) OnPing(e *handler.CommandEvent) error {
	return e.CreateMessage(res.Create("Pong!"))
}
