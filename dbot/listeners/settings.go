package listeners

import (
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/events"

	"github.com/KittyBot-Org/KittyBotGo/dbot"
)

func Settings(b *dbot.Bot) bot.EventListener {
	return bot.NewListenerFunc(func(event bot.Event) {
		switch e := event.(type) {
		case *events.GuildJoin:
			if err := b.DB.GuildSettings().CreateIfNotExist(e.Guild.ID); err != nil {
				b.Logger.Error("Failed to create guild settings: ", err)
			}

		case *events.GuildLeave:
			if err := b.DB.GuildSettings().Delete(e.Guild.ID); err != nil {
				b.Logger.Error("Failed to delete guild settings: ", err)
			}

		case *events.GuildReady:
			if err := b.DB.GuildSettings().CreateIfNotExist(e.Guild.ID); err != nil {
				b.Logger.Error("Failed to create guild settings: ", err)
			}
		}
	})
}
