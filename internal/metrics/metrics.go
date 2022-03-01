package metrics

import (
	"net/http"

	"github.com/KittyBot-Org/KittyBotGo/internal/types"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	ShardCounter = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "bot_shards",
		Help: "The total number of shards the bot has",
	})

	GuildCounter = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "bot_guilds",
		Help: "The total number of guilds the bot is in",
	})

	UserCounter = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "bot_users",
		Help: "The total number of users the bot serves",
	})

	CommandsHandledCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "bot_commands_handled",
		Help: "The total number of commands handled by the bot",
	})

	ComponentsHandledCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "bot_components_handled",
		Help: "The total number of components handled by the bot",
	})
)

func Setup(b *types.Bot) {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	go func() {
		if err := http.ListenAndServe(":2112", mux); err != nil {
			b.Logger.Error("Failed to start metrics server", err)
		}
	}()
}
