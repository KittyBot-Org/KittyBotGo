package commands

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"

	"github.com/KittyBot-Org/KittyBotGo/service/bot/res"
)

func (c *commands) OnHasPlayer(next handler.Handler) handler.Handler {
	return func(e *handler.InteractionEvent) error {
		player := c.Lavalink.ExistingPlayer(*e.GuildID())
		if player == nil {
			return e.Respond(discord.InteractionResponseTypeCreateMessage, res.CreateError("No player found"))
		}
		return next(e)
	}
}
