package metrics

import (
	"github.com/DisgoOrg/disgo/core"
	"github.com/DisgoOrg/disgo/core/events"
	"github.com/KittyBot-Org/KittyBotGo/internal/bot/metrics"
	"github.com/KittyBot-Org/KittyBotGo/internal/bot/types"
)

var Module = module{}

type module struct{}

func (m module) OnEvent(b *types.Bot, event core.Event) {
	switch e := event.(type) {
	case *events.GuildsReadyEvent:
		b.Logger.Info("Guilds ready, setting counter")
		metrics.GuildCounter.Set(float64(len(e.Bot().Caches.Guilds().Cache())))

	case *events.GuildJoinEvent:
		metrics.GuildCounter.Inc()

	case *events.GuildLeaveEvent:
		metrics.GuildCounter.Dec()
	}
}
