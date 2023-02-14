package commands

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/handler"

	"github.com/KittyBot-Org/KittyBotGo/service/bot/res"
)

func (c *Cmds) OnHasPlayer(next handler.Handler) handler.Handler {
	return func(e *events.InteractionCreate) {
		ok, err := c.Database.HasPlayer(*e.GuildID())
		if err != nil {
			_ = e.Respond(discord.InteractionResponseTypeCreateMessage, res.CreateErr("Error checking player", err))
			return
		}
		if !ok {
			_ = e.Respond(discord.InteractionResponseTypeCreateMessage, res.CreateError("No player found"))
			return
		}
		next(e)
	}
}
