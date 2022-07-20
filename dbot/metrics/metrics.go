package metrics

import (
	"github.com/KittyBot-Org/KittyBotGo/dbot"

	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	ShardCounter = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "kittybot_shard_count",
		Help: "The total number of shards kittybot has",
	})

	ShardStatus = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "kittybot_shard_status",
		Help: "The total number of shards kittybot has",
	})

	GuildCounter = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "kittybot_guild_count",
		Help: "The total number of guilds kittybot is in",
	})

	UserCounter = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "kittybot_user_count",
		Help: "The total number of users kittybot serves",
	})

	AudioPlayerCounter = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "kittybot_audio_player_count",
		Help: "The total number of active audio players kittybot has",
	})

	CommandsHandledCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "kittybot_commands_handled",
		Help: "The total number of commands handled by the dbot",
	})

	ComponentsHandledCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "kittybot_components_handled",
		Help: "The total number of components handled by the dbot",
	})
)

func Setup(b *dbot.Bot) {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	go func() {
		if err := http.ListenAndServe(b.Config.MetricsAddress, mux); err != nil {
			b.Logger.Error("Failed to start metrics server: ", err)
		}
	}()
}
