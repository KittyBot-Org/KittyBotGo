package routes

import (
	"net/http"

	"github.com/KittyBot-Org/KittyBotGo/backend"
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
