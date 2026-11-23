package commands

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/json/v2"

	"github.com/kittybot-org/kittybotgo/service/bot"
	"github.com/kittybot-org/kittybotgo/service/bot/res"
)

func (c *commands) OnHasPlayer(next handler.Handler) handler.Handler {
	return func(e *handler.InteractionEvent) error {
		player := c.Lavalink.ExistingPlayer(*e.GuildID())
		if player == nil || player.Node == nil {
			return e.Respond(discord.InteractionResponseTypeCreateMessage, res.CreateError("No player found"))
		}
		return next(e)
	}
}

func (c *commands) OnDJ(next handler.Handler) handler.Handler {
	return func(e *handler.InteractionEvent) error {
		settings, err := c.Database.GetGuild(e.Ctx, *e.GuildID())
		if err != nil || settings.DJRole() == 0 {
			return next(e)
		}
		if member := e.Member(); member != nil {
			if member.Permissions.Has(discord.PermissionAdministrator) {
				return next(e)
			}
			for _, roleID := range member.RoleIDs {
				if roleID == settings.DJRole() {
					return next(e)
				}
			}
		}
		player := c.Lavalink.ExistingPlayer(*e.GuildID())
		if player != nil && player.Track != nil {
			var userData bot.UserData
			if err := json.Unmarshal([]byte(player.Track.UserData), &userData); err == nil && userData.Requester == e.User().ID {
				return next(e)
			}
		}
		return e.Respond(discord.InteractionResponseTypeCreateMessage, res.CreateError("You are not the song requester or the DJ"))
	}
}
