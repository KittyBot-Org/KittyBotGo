package routes

import (
	"encoding/json"
	"net/http"

	"github.com/KittyBot-Org/KittyBotGo/internal/backend/types"
	"github.com/gorilla/mux"
)

func Handler(b *types.Backend) http.Handler {
	router := mux.NewRouter()

	router.HandleFunc("/info", InfoHandler(b)).Methods(http.MethodGet)
	router.HandleFunc("/health_check", HealthCheckHandler).Methods(http.MethodGet)
	router.HandleFunc("/votes/{botlist}", VotesHandler(b)).Methods(http.MethodPost)
	router.HandleFunc("/bot_invite", BotInviteHandler(b)).Methods(http.MethodGet)
	router.HandleFunc("/guild_invite", GuildInviteHandler(b)).Methods(http.MethodGet)
	router.HandleFunc("/commands", CommandsHandler(b)).Methods(http.MethodGet)

	return router
}

func HealthCheckHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("alive"))
}

func CommandsHandler(b *types.Backend) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(b.Commands); err != nil {
			b.Logger.Error(err)
		}
	}
}
