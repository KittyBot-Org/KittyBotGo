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
			likedSongsCommand,
		},
	}
	handlers.Command("/ping", handlers.OnPing)

	handlers.Route("/player", func(r handler.Router) {
		r.Command("/play", handlers.OnPlayerPlay)
		r.Autocomplete("/play", handlers.OnPlayerPlayAutocomplete)
		r.Group(func(r handler.Router) {
			r.Use(handlers.OnHasPlayer)
			r.Command("/status", handlers.OnPlayerStatus)
			r.Command("/pause", handlers.OnPlayerPause)
			r.Command("/resume", handlers.OnPlayerResume)
			r.Command("/stop", handlers.OnPlayerStop)
			r.Command("/next", handlers.OnPlayerNext)
			r.Command("/previous", handlers.OnPlayerPrevious)
			r.Command("/volume", handlers.OnPlayerVolume)
			r.Command("/bass-boost", handlers.OnPlayerBassBoost)

			r.Component("/previous", handlers.OnPlayerPreviousButton)
			r.Component("/pause_play", handlers.OnPlayerPlayPauseButton)
			r.Component("/next", handlers.OnPlayerNextButton)
			r.Component("/stop", handlers.OnPlayerStopButton)
		})
	})
	handlers.Route("/queue", func(r handler.Router) {
		r.Use(handlers.OnHasPlayer)
		r.Command("/clear", handlers.OnQueueClear)
		r.Command("/remove", handlers.OnQueueRemove)
		r.Autocomplete("/remove", handlers.OnQueueAutocomplete)
		r.Command("/shuffle", handlers.OnQueueShuffle)
		r.Command("/show", handlers.OnQueueShow)
		r.Command("/type", handlers.OnQueueType)
	})
	handlers.Route("/history", func(r handler.Router) {
		r.Use(handlers.OnHasPlayer)
		r.Command("/clear", handlers.OnHistoryClear)
		r.Command("/show", handlers.OnHistoryShow)
	})
	handlers.Route("/playlists", func(r handler.Router) {
		r.Command("/list", handlers.OnPlaylistsList)
		r.Command("/show", handlers.OnPlaylistShow)
		r.Autocomplete("/show", handlers.OnPlaylistAutocomplete)
		r.Command("/play", handlers.OnPlaylistPlay)
		r.Autocomplete("/play", handlers.OnPlaylistAutocomplete)
		r.Command("/create", handlers.OnPlaylistCreate)
		r.Command("/delete", handlers.OnPlaylistDelete)
		r.Autocomplete("/delete", handlers.OnPlaylistAutocomplete)
		r.Command("/add", handlers.OnPlaylistAdd)
		r.Autocomplete("/add", handlers.OnPlaylistAutocomplete)
		r.Command("/remove", handlers.OnPlaylistRemove)
		r.Autocomplete("/remove", handlers.OnPlaylistRemoveAutocomplete)
	})
	handlers.Route("/liked-songs", func(r handler.Router) {
		r.Command("/show", handlers.OnLikedSongsShow)
		//r.Command("/add", handlers.OnLikedSongsAdd)
		r.Command("/remove", handlers.OnLikedSongsRemove)
		r.Autocomplete("/remove", handlers.OnLikedSongsAutocomplete)
		r.Command("/clear", handlers.OnLikedSongsClear)
		r.Component("/add", handlers.OnLikedSongsAddButton)
	})

	return handlers
}

type Handlers struct {
	*bot.Bot
	handler.Router
	Commands []discord.ApplicationCommandCreate
}
