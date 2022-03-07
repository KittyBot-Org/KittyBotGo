package routes

import (
	"net/http"

	"github.com/KittyBot-Org/KittyBotGo/internal/backend/types"
	"github.com/gorilla/mux"
)

func Handler(b *types.Backend) http.Handler {
	router := mux.NewRouter()

	router.HandleFunc("/info", InfoHandler(b)).Methods(http.MethodGet)

	router.HandleFunc("/votes/{botlist}", VotesHandler(b)).Methods(http.MethodPost)

	router.HandleFunc("/bot_invite", BotInviteHandler(b)).Methods(http.MethodGet)
	router.HandleFunc("/guild_invite", GuildInviteHandler(b)).Methods(http.MethodGet)

	return router
}
