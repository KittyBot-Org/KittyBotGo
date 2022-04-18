package routes

import (
	"github.com/KittyBot-Org/KittyBotGo/internal/backend"
	"net/http"
)

func BotInviteHandler(b *backend.Backend) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, b.Config.BotInvite, http.StatusTemporaryRedirect)
	}
}

func GuildInviteHandler(b *backend.Backend) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, b.Config.GuildInvite, http.StatusTemporaryRedirect)
	}
}
