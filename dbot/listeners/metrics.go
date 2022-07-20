package listeners

import (
	"github.com/KittyBot-Org/KittyBotGo/dbot"
	"github.com/KittyBot-Org/KittyBotGo/dbot/metrics"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/events"
)

func Metrics(b *dbot.Bot) bot.EventListener {
	return bot.NewListenerFunc(func(event bot.Event) {
		switch e := event.(type) {
		case *events.GuildsReady:
			b.Logger.Info("Guilds ready, setting counter")
			metrics.GuildCounter.Set(float64(len(e.Client().Caches().Guilds().All())))

		case *events.GuildJoin:
			metrics.GuildCounter.Inc()

		case *events.GuildLeave:
			metrics.GuildCounter.Dec()
		}
	})
}
