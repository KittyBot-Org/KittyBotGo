package routes

import (
	"github.com/KittyBot-Org/KittyBotGo/internal/bend"
	"net/http"
)

func BotInviteHandler(b *bend.Backend) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, b.Config.BotInvite, http.StatusTemporaryRedirect)
	}
}

func GuildInviteHandler(b *bend.Backend) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, b.Config.GuildInvite, http.StatusTemporaryRedirect)
	}
}
