package commands

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/disgo/handler/middleware"

	"github.com/kittybot-org/kittybotgo/service/bot"
)

var Commands = []discord.ApplicationCommandCreate{
	pingCommand,
	playerCommand,
	queueCommand,
	historyCommand,
	playlistsCommand,
	likedSongsCommand,
	settingsCommand,
	adminCommand,
	tagsCommand,
	notificationCommand,
	nekoCommand,
	utilityCommand,
	infoCommand,
}

type commands struct {
	*bot.Bot
}

func New(b *bot.Bot) handler.Router {
	cmds := &commands{b}

	router := handler.New()
	router.Use(middleware.Go)
	router.SlashCommand("/ping", cmds.OnPing)
	router.Route("/music", func(r handler.Router) {
		r.SlashCommand("/play", cmds.OnPlayerPlay)
		r.Autocomplete("/play", cmds.OnPlayerPlayAutocomplete)
		r.Group(func(r handler.Router) {
			r.Use(cmds.OnHasPlayer)
			r.Use(cmds.OnDJ)
			r.SlashCommand("/status", cmds.OnPlayerStatus)
			r.SlashCommand("/pause", cmds.OnPlayerPause)
			r.SlashCommand("/resume", cmds.OnPlayerResume)
			r.SlashCommand("/stop", cmds.OnPlayerStop)
			r.SlashCommand("/next", cmds.OnPlayerNext)
			r.SlashCommand("/previous", cmds.OnPlayerPrevious)
			r.SlashCommand("/volume", cmds.OnPlayerVolume)
			r.SlashCommand("/bass-boost", cmds.OnPlayerBassBoost)
			r.SlashCommand("/seek", cmds.OnPlayerSeek)
			r.SlashCommand("/forward", cmds.OnPlayerForward)
			r.SlashCommand("/rewind", cmds.OnPlayerRewind)
			r.Route("/filter", func(r handler.Router) {
				r.SlashCommand("/equalizer", cmds.OnFilterEqualizer)
				r.SlashCommand("/karaoke", cmds.OnFilterKaraoke)
				r.SlashCommand("/timescale", cmds.OnFilterTimescale)
				r.SlashCommand("/tremolo", cmds.OnFilterTremolo)
				r.SlashCommand("/vibrato", cmds.OnFilterVibrato)
				r.SlashCommand("/rotation", cmds.OnFilterRotation)
				r.SlashCommand("/distortion", cmds.OnFilterDistortion)
				r.SlashCommand("/channel-mix", cmds.OnFilterChannelMix)
				r.SlashCommand("/low-pass", cmds.OnFilterLowPass)
				r.SlashCommand("/clear", cmds.OnFilterClear)
			})

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
	router.Route("/settings", func(r handler.Router) {
		r.SlashCommand("/list", cmds.OnSettingsList)
		r.SlashCommand("/joinmessage", cmds.OnSettingsJoin)
		r.SlashCommand("/leavemessage", cmds.OnSettingsLeave)
		r.SlashCommand("/announcementchannel", cmds.OnSettingsAnnouncement)
		r.SlashCommand("/djrole", cmds.OnSettingsDJRole)
		r.SlashCommand("/nsfw", cmds.OnSettingsNSFW)
		r.SlashCommand("/rolesaver", cmds.OnSettingsRoleSaver)
	})
	router.Route("/admin", func(r handler.Router) {
		r.Route("/ban", func(r handler.Router) {
			r.SlashCommand("/add", cmds.OnAdminBanAdd)
			r.SlashCommand("/remove", cmds.OnAdminBanRemove)
			r.SlashCommand("/list", cmds.OnAdminBanList)
		})
		r.SlashCommand("/kick", cmds.OnAdminKick)
	})
	router.Route("/tags", func(r handler.Router) {
		r.SlashCommand("/create", cmds.OnTagCreate)
		r.SlashCommand("/delete", cmds.OnTagDelete)
		r.Autocomplete("/delete", cmds.OnTagAutocomplete)
		r.SlashCommand("/edit", cmds.OnTagEdit)
		r.Autocomplete("/edit", cmds.OnTagAutocomplete)
		r.SlashCommand("/info", cmds.OnTagInfo)
		r.Autocomplete("/info", cmds.OnTagAutocomplete)
		r.SlashCommand("/list", cmds.OnTagList)
		r.SlashCommand("/search", cmds.OnTagSearch)
		r.SlashCommand("/publish", cmds.OnTagPublish)
		r.Autocomplete("/publish", cmds.OnTagAutocomplete)
		r.SlashCommand("/unpublish", cmds.OnTagUnpublish)
		r.Autocomplete("/unpublish", cmds.OnTagAutocomplete)
		r.SlashCommand("/send", cmds.OnTagSend)
		r.Autocomplete("/send", cmds.OnTagAutocomplete)
	})
	router.Route("/notification", func(r handler.Router) {
		r.SlashCommand("/create", cmds.OnNotificationCreate)
		r.SlashCommand("/delete", cmds.OnNotificationDelete)
		r.SlashCommand("/list", cmds.OnNotificationList)
	})
	router.Route("/neko", func(r handler.Router) {
		for _, name := range []string{"bite", "cuddle", "feed", "hug", "kiss", "lick", "neko", "pat", "poke", "slap", "tickle"} {
			r.SlashCommand("/"+name, cmds.OnNeko)
		}
	})
	router.Route("/utility", func(r handler.Router) {
		r.SlashCommand("/hastebin", cmds.OnHastebin)
		r.SlashCommand("/translate", cmds.OnTranslate)
		r.SlashCommand("/steal-emote", cmds.OnStealEmote)
	})
	router.Route("/info", func(r handler.Router) {
		r.SlashCommand("/help", cmds.OnInfoHelp)
		r.SlashCommand("/info", cmds.OnInfoInfo)
		r.SlashCommand("/avatar", cmds.OnInfoAvatar)
		r.SlashCommand("/servericon", cmds.OnInfoServerIcon)
		r.SlashCommand("/serverbanner", cmds.OnInfoServerBanner)
		r.SlashCommand("/uptime", cmds.OnInfoUptime)
		r.SlashCommand("/vote", cmds.OnInfoVote)
		r.SlashCommand("/privacy", cmds.OnInfoPrivacy)
		r.SlashCommand("/commands", cmds.OnInfoCommands)
		r.SlashCommand("/dashboard", cmds.OnInfoDashboard)
	})

	return router
}
