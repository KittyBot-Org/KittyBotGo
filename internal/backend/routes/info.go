package routes

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/DisgoOrg/disgo/info"
	"github.com/KittyBot-Org/KittyBotGo/internal/backend/types"
	"github.com/pkg/errors"
	"github.com/prometheus/common/model"
)

type Stats struct {
	GuildCount       int      `json:"guild_count"`
	UserCount        int      `json:"user_count"`
	ShardCount       int      `json:"shard_count"`
	ShardStatus      []string `json:"shard_status"`
	AudioPlayerCount int      `json:"audio_player_count"`
}

func getMetric(ctx context.Context, b *types.Backend, query string) (int, error) {
	result, warnings, err := b.PrometheusAPI.Query(ctx, query, time.Time{})
	if err != nil {
		return 0, err
	}
	if len(warnings) > 0 {
		b.Logger.Warnf("Warnings while running query %s: %s", query, warnings)
	}
	vectorResult, ok := result.(model.Vector)
	if !ok {
		return 0, errors.Errorf("unexpected result type %T for query: %s", result, query)
	}
	return int(vectorResult[0].Value), nil
}

func getMetrics(ctx context.Context, b *types.Backend) (stats *Stats, err error) {
	if stats.GuildCount, err = getMetric(ctx, b, "kittybot_guild_count"); err != nil {
		return nil, err
	}
	if stats.UserCount, err = getMetric(ctx, b, "kittybot_user_count"); err != nil {
		return nil, err
	}
	if stats.ShardCount, err = getMetric(ctx, b, "kittybot_shard_count"); err != nil {
		return nil, err
	}
	if stats.GuildCount, err = getMetric(ctx, b, "kittybot_audio_player_count"); err != nil {
		return nil, err
	}
	return
}

func InfoHandler(b *types.Backend) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stats, err := getMetrics(r.Context(), b)
		if err != nil {
			b.Logger.Error("failed to request metrics: ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if err = json.NewEncoder(w).Encode(struct {
			BotVersion   string `json:"bot_version"`
			DisgoVersion string `json:"disgo_version"`
			CommandCount int    `json:"command_count"`
			Stats
		}{
			BotVersion:   b.Version,
			DisgoVersion: info.Version,
			CommandCount: len(b.Commands),
			Stats:        *stats,
		}); err != nil {
			b.Logger.Error(err)
		}
	}
}
