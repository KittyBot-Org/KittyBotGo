package commands

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"

	"github.com/KittyBot-Org/KittyBotGo/service/bot/res"
)

var pingCommand = discord.SlashCommandCreate{
	Name:        "ping",
	Description: "Ping the bot",
}

func (c *commands) OnPing(_ discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	return e.CreateMessage(res.Create("Pong!"))
}
