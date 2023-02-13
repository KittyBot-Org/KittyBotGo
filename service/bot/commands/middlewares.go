package commands

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/handler"

	"github.com/KittyBot-Org/KittyBotGo/service/bot/res"
)

func (h *Cmds) OnHasPlayer(next handler.Handler) handler.Handler {
	return func(e *events.InteractionCreate) {
		if !h.HasPlayer(*e.GuildID()) {
			_ = e.Respond(discord.InteractionResponseTypeCreateMessage, res.CreateError("No player found"))
			return
		}
		next(e)
	}
}
