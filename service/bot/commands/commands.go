package commands

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/disgo/handler/middleware"

	"github.com/KittyBot-Org/KittyBotGo/service/bot"
)

var Commands = []discord.ApplicationCommandCreate{
	pingCommand,
	playerCommand,
	queueCommand,
	historyCommand,
	playlistsCommand,
	likedSongsCommand,
}

type commands struct {
	*bot.Bot
}

func New(b *bot.Bot) handler.Router {
	cmds := &commands{b}

	router := handler.New()
	router.Use(middleware.Go)
	router.SlashCommand("/ping", cmds.OnPing)
	router.Route("/player", func(r handler.Router) {
		r.SlashCommand("/play", cmds.OnPlayerPlay)
		r.Autocomplete("/play", cmds.OnPlayerPlayAutocomplete)
		r.Group(func(r handler.Router) {
			r.Use(cmds.OnHasPlayer)
			r.SlashCommand("/status", cmds.OnPlayerStatus)
			r.SlashCommand("/pause", cmds.OnPlayerPause)
			r.SlashCommand("/resume", cmds.OnPlayerResume)
			r.SlashCommand("/stop", cmds.OnPlayerStop)
			r.SlashCommand("/next", cmds.OnPlayerNext)
			r.SlashCommand("/previous", cmds.OnPlayerPrevious)
			r.SlashCommand("/volume", cmds.OnPlayerVolume)
			r.SlashCommand("/bass-boost", cmds.OnPlayerBassBoost)
			r.SlashCommand("/seek", cmds.OnPlayerSeek)

			r.Component("/previous", cmds.OnPlayerPreviousButton)
			r.Component("/pause_play", cmds.OnPlayerPlayPauseButton)
			r.Component("/next", cmds.OnPlayerNextButton)
			r.Component("/stop", cmds.OnPlayerStopButton)
		})
	})
	router.Route("/queue", func(r handler.Router) {
		r.Use(cmds.OnHasPlayer)
		r.SlashCommand("/clear", cmds.OnQueueClear)
		r.SlashCommand("/remove", cmds.OnQueueRemove)
		r.Autocomplete("/remove", cmds.OnQueueAutocomplete)
		r.SlashCommand("/shuffle", cmds.OnQueueShuffle)
		r.SlashCommand("/show", cmds.OnQueueShow)
		r.SlashCommand("/type", cmds.OnQueueType)
	})
	router.Route("/history", func(r handler.Router) {
		r.Use(cmds.OnHasPlayer)
		r.SlashCommand("/clear", cmds.OnHistoryClear)
		r.SlashCommand("/show", cmds.OnHistoryShow)
	})
	router.Route("/playlists", func(r handler.Router) {
		r.SlashCommand("/list", cmds.OnPlaylistsList)
		r.SlashCommand("/show", cmds.OnPlaylistShow)
		r.Autocomplete("/show", cmds.OnPlaylistAutocomplete)
		r.SlashCommand("/play", cmds.OnPlaylistPlay)
		r.Autocomplete("/play", cmds.OnPlaylistAutocomplete)
		r.SlashCommand("/create", cmds.OnPlaylistCreate)
		r.SlashCommand("/delete", cmds.OnPlaylistDelete)
		r.Autocomplete("/delete", cmds.OnPlaylistAutocomplete)
		r.SlashCommand("/add", cmds.OnPlaylistAdd)
		r.Autocomplete("/add", cmds.OnPlaylistAutocomplete)
		r.SlashCommand("/remove", cmds.OnPlaylistRemove)
		r.Autocomplete("/remove", cmds.OnPlaylistRemoveAutocomplete)
	})
	router.Route("/liked-songs", func(r handler.Router) {
		r.SlashCommand("/show", cmds.OnLikedSongsShow)
		// r.SlashCommand("/add", cmds.OnLikedSongsAdd)
		r.SlashCommand("/remove", cmds.OnLikedSongsRemove)
		r.Autocomplete("/remove", cmds.OnLikedSongsAutocomplete)
		r.SlashCommand("/clear", cmds.OnLikedSongsClear)
		r.Component("/add", cmds.OnLikedSongsAddButton)
	})

	return router
}
