package metrics

import (
	"github.com/KittyBot-Org/KittyBotGo/internal/bot/metrics"
	"github.com/KittyBot-Org/KittyBotGo/internal/bot/types"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/events"
)

var Module = module{}

type module struct{}

func (m module) OnEvent(b *types.Bot, event bot.Event) {
	switch e := event.(type) {
	case *events.GuildsReadyEvent:
		b.Logger.Info("Guilds ready, setting counter")
		metrics.GuildCounter.Set(float64(len(e.Client().Caches().Guilds().All())))

	case *events.GuildJoinEvent:
		metrics.GuildCounter.Inc()

	case *events.GuildLeaveEvent:
		metrics.GuildCounter.Dec()
	}
}
