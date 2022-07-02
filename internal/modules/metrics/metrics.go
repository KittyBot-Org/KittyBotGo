package metrics

import (
	"github.com/KittyBot-Org/KittyBotGo/internal/kbot"
	"github.com/KittyBot-Org/KittyBotGo/internal/metrics"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/events"
)

var Module = module{}

type module struct{}

func (m module) OnEvent(b *kbot.Bot, event bot.Event) {
	switch e := event.(type) {
	case *events.GuildsReady:
		b.Logger.Info("Guilds ready, setting counter")
		metrics.GuildCounter.Set(float64(len(e.Client().Caches().Guilds().All())))

	case *events.GuildJoin:
		metrics.GuildCounter.Inc()

	case *events.GuildLeave:
		metrics.GuildCounter.Dec()
	}
}
