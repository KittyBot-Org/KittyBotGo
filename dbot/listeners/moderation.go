package listeners

import (
	"github.com/KittyBot-Org/KittyBotGo/dbot"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/events"
)

func Moderation(b *dbot.Bot) bot.EventListener {
	return bot.NewListenerFunc(func(e *events.AutoModerationActionExecution) {

	})
}
