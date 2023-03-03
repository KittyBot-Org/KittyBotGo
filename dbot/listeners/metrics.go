package listeners

import (
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"

	"github.com/KittyBot-Org/KittyBotGo/dbot"
	"github.com/KittyBot-Org/KittyBotGo/dbot/metrics"
)

func Metrics(b *dbot.Bot) bot.EventListener {
	return bot.NewListenerFunc(func(event bot.Event) {
		switch e := event.(type) {
		case *events.GuildsReady:
			b.Logger.Info("Guilds ready, setting counter")

			var count float64
			e.Client().Caches().GuildsForEach(func(guild discord.Guild) {
				count++
			})
			metrics.GuildCounter.Set(count)

		case *events.GuildJoin:
			metrics.GuildCounter.Inc()

		case *events.GuildLeave:
			metrics.GuildCounter.Dec()
		}
	})
}
