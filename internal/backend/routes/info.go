package routes

import (
	"encoding/json"
	"net/http"

	"github.com/DisgoOrg/disgo/info"
	"github.com/KittyBot-Org/KittyBotGo/internal/backend/types"
)

func InfoHandler(b *types.Backend) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(struct {
			BotVersion   string `json:"bot_version"`
			DisgoVersion string `json:"disgo_version"`
		}{
			BotVersion:   b.Version,
			DisgoVersion: info.Version,
		}); err != nil {
			b.Logger.Error(err)
		}
	}
}
