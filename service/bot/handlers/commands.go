package handlers

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"

	"github.com/KittyBot-Org/KittyBotGo/service/bot"
)

func New(b *bot.Bot) *Handlers {
	handlers := &Handlers{
		Bot:    b,
		Router: handler.New(),
		Commands: []discord.ApplicationCommandCreate{
			pingCommand,
			playerCommand,
			queueCommand,
			historyCommand,
			playlistsCommand,
		},
	}
	handlers.HandleCommand("/ping", handlers.OnPing)

	handlers.Route("/player", func(r handler.Router) {
		r.HandleCommand("/play", handlers.OnPlayerPlay)
		r.Group(func(r handler.Router) {
			r.Use(handlers.OnHasPlayer)
			r.HandleCommand("/status", handlers.OnPlayerStatus)
			r.HandleCommand("/pause", handlers.OnPlayerPause)
			r.HandleCommand("/resume", handlers.OnPlayerResume)
			r.HandleCommand("/stop", handlers.OnPlayerStop)
			r.HandleCommand("/next", handlers.OnPlayerNext)
			r.HandleCommand("/previous", handlers.OnPlayerPrevious)
			r.HandleCommand("/volume", handlers.OnPlayerVolume)
			r.HandleCommand("/bass-boost", handlers.OnPlayerBassBoost)
		})
	})
	handlers.Route("/queue", func(r handler.Router) {
		r.Use(handlers.OnHasPlayer)
		r.HandleCommand("/clear", handlers.OnQueueClear)
		r.HandleCommand("/remove", handlers.OnQueueRemove)
		r.HandleAutocomplete("/remove", handlers.OnQueueAutocomplete)
		r.HandleCommand("/shuffle", handlers.OnQueueShuffle)
		r.HandleCommand("/show", handlers.OnQueueShow)
		r.HandleCommand("/type", handlers.OnQueueType)
	})
	handlers.Route("/history", func(r handler.Router) {
		r.Use(handlers.OnHasPlayer)
		r.HandleCommand("/clear", handlers.OnHistoryClear)
		r.HandleCommand("/show", handlers.OnHistoryShow)
	})
	handlers.Route("/playlists", func(r handler.Router) {
		r.HandleCommand("/list", handlers.OnPlaylistsList)
		r.HandleCommand("/show", handlers.OnPlaylistShow)
		r.HandleAutocomplete("/show", handlers.OnPlaylistAutocomplete)
		r.HandleCommand("/create", handlers.OnPlaylistCreate)
		r.HandleCommand("/delete", handlers.OnPlaylistDelete)
		r.HandleAutocomplete("/delete", handlers.OnPlaylistAutocomplete)
		//r.HandleCommand("/add", handlers.OnPlaylistsAdd)
		//r.HandleAutocomplete("/add", handlers.OnPlaylistsPlaylist)
	})

	return handlers
}

type Handlers struct {
	*bot.Bot
	handler.Router
	Commands []discord.ApplicationCommandCreate
}
