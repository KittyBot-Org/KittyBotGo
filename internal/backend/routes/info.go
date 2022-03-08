package routes

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/DisgoOrg/disgo/info"
	"github.com/KittyBot-Org/KittyBotGo/internal/backend/types"
)

func InfoHandler(b *types.Backend) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		result, warnings, err := b.PrometheusAPI.Query(r.Context(), "kittybot_guilds", time.Now())
		if err != nil {
			b.Logger.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if len(warnings) > 0 {
			b.Logger.Warn("Warnings while querying prometheus api: ", warnings)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if err = json.NewEncoder(w).Encode(struct {
			BotVersion   string `json:"bot_version"`
			DisgoVersion string `json:"disgo_version"`
			CommandCount int    `json:"command_count"`
			GuildCount   string `json:"guild_count"`
			UserCount    string `json:"user_count"`
		}{
			BotVersion:   b.Version,
			DisgoVersion: info.Version,
			CommandCount: len(b.Commands),
			GuildCount:   result.String(),
			UserCount:    "",
		}); err != nil {
			b.Logger.Error(err)
		}
	}
}
