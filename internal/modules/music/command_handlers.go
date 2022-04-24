package music

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/KittyBot-Org/KittyBotGo/internal/kbot"
	"github.com/KittyBot-Org/KittyBotGo/internal/responses"
	"github.com/disgoorg/source-extensions-plugin"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgolink/lavalink"
	"github.com/disgoorg/snowflake"
	"github.com/disgoorg/utils/paginator"
	"golang.org/x/text/message"
)

var (
	urlPattern    = regexp.MustCompile("^https?://[-a-zA-Z0-9+&@#/%?=~_|!:,.;]*[-a-zA-Z0-9+&@#/%=~_|]?")
	searchPattern = regexp.MustCompile(`^(.{2})search:(.+)`)

	bassBoost = &lavalink.Equalizer{
		0:  0.2,
		1:  0.15,
		2:  0.1,
		3:  0.05,
		4:  0.0,
		5:  -0.05,
		6:  -0.1,
		7:  -0.1,
		8:  -0.1,
		9:  -0.1,
		10: -0.1,
		11: -0.1,
		12: -0.1,
		13: -0.1,
		14: -0.1,
	}
)

func playHandler(b *kbot.Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) error {
	data := e.SlashCommandInteractionData()

	var voiceChannelID snowflake.Snowflake
	if voiceState, ok := b.Client.Caches().VoiceStates().Get(*e.GuildID(), e.User().ID); !ok || voiceState.ChannelID == nil {
		return e.CreateMessage(responses.CreateErrorf(p, "modules.music.not.in.voice"))
	} else {
		voiceChannelID = *voiceState.ChannelID
	}

	channel, _ := e.GuildChannel()
	if perms := b.Client.Caches().GetMemberPermissionsInChannel(channel, e.Member().Member); perms.Missing(discord.PermissionVoiceConnect) {
		return e.CreateMessage(responses.CreateErrorf(p, "modules.music.missing.perms"))
	}

	if voiceState, ok := b.Client.Caches().VoiceStates().Get(*e.GuildID(), b.Client.ID()); !ok || voiceState.ChannelID == nil || *voiceState.ChannelID != voiceChannelID {
		if err := b.Client.Connect(context.TODO(), *e.GuildID(), voiceChannelID); err != nil {
			return e.CreateMessage(responses.CreateErrorf(p, "modules.music.connect.error"))
		}
	}

	query := data.String("query")
	if searchProvider, ok := data.OptString("searchProvider"); ok {
		query = lavalink.SearchType(searchProvider).Apply(query)
	} else {
		if !urlPattern.MatchString(query) && !searchPattern.MatchString(query) {
			query = lavalink.SearchTypeYoutube.Apply(query)
		}
	}

	if err := e.DeferCreateMessage(false); err != nil {
		return err
	}
	err := b.Lavalink.BestRestClient().LoadItemHandler(context.TODO(), query, lavalink.NewResultHandler(
		func(track lavalink.AudioTrack) {
			if err := b.DB.PlayHistory().Add(e.User().ID, query, track.Info().Title); err != nil {
				b.Logger.Error("Failed to add track to play history: ", err)
			}
			playAndQueue(b, p, e.BaseInteraction, track)
		},
		func(playlist lavalink.AudioPlaylist) {
			if err := b.DB.PlayHistory().Add(e.User().ID, query, playlist.Name()); err != nil {
				b.Logger.Error("Failed to add track to play history: ", err)
			}
			playAndQueue(b, p, e.BaseInteraction, playlist.Tracks()...)
		},
		func(tracks []lavalink.AudioTrack) {
			if err := b.DB.PlayHistory().Add(e.User().ID, query, data.String("query")); err != nil {
				b.Logger.Error("Failed to add track to play history: ", err)
			}
			giveSearchSelection(b, p, e, tracks)
		},
		func() {
			if _, err := e.Client().Rest().UpdateInteractionResponse(e.ApplicationID(), e.Token(), responses.UpdateErrorf(p, "modules.music.commands.play.no.results")); err != nil {
				b.Logger.Error("Failed to update not found message: ", err)
			}
		},
		func(ex lavalink.FriendlyException) {
			if _, err := e.Client().Rest().UpdateInteractionResponse(e.ApplicationID(), e.Token(), responses.UpdateErrorf(p, "modules.music.commands.play.error", ex.Message)); err != nil {
				b.Logger.Error("Failed to update error message: ", err)
			}
		},
	))
	if err != nil {
		if _, err = e.Client().Rest().UpdateInteractionResponse(e.ApplicationID(), e.Token(), responses.UpdateErrorf(p, "modules.music.commands.play.error")); err != nil {
			b.Logger.Error("Failed to update error message: ", err)
		}
	}
	return err
}

func playAndQueue(b *kbot.Bot, p *message.Printer, i discord.BaseInteraction, tracks ...lavalink.AudioTrack) {
	player := b.MusicPlayers.Get(*i.GuildID())
	if player == nil {
		player = b.MusicPlayers.New(*i.GuildID(), kbot.PlayerTypeMusic, kbot.LoopingTypeOff)
		b.MusicPlayers.Add(player)
	}
	var voiceChannelID snowflake.Snowflake
	if voiceState, ok := b.Client.Caches().VoiceStates().Get(*i.GuildID(), i.User().ID); !ok || voiceState.ChannelID == nil {
		if _, err := b.Client.Rest().UpdateInteractionResponse(i.ApplicationID(), i.Token(), responses.UpdateErrorComponentsf(p, "modules.music.not.in.voice", nil)); err != nil {
			b.Logger.Error("Failed to update error message: ", err)
		}
		return
	} else {
		voiceChannelID = *voiceState.ChannelID
	}

	if voiceState, ok := b.Client.Caches().VoiceStates().Get(*i.GuildID(), b.Client.ID()); !ok || voiceState.ChannelID == nil || *voiceState.ChannelID != voiceChannelID {
		if err := b.Client.Connect(context.TODO(), *i.GuildID(), voiceChannelID); err != nil {
			if _, err = b.Client.Rest().UpdateInteractionResponse(i.ApplicationID(), i.Token(), responses.UpdateErrorComponentsf(p, "modules.music.connect.error", nil)); err != nil {
				b.Logger.Error("Failed to update error message: ", err)
			}
			return
		}
	}

	for ii := range tracks {
		tracks[ii].SetUserData(kbot.AudioTrackData{
			Requester: i.User().ID,
		})
	}

	if player.PlayingTrack() == nil {
		track := tracks[0]
		if len(tracks) > 0 {
			tracks = tracks[1:]
		}
		if err := player.Play(track); err != nil {
			if _, err = b.Client.Rest().UpdateInteractionResponse(i.ApplicationID(), i.Token(), responses.UpdateErrorComponentsf(p, "modules.music.commands.play.error", nil)); err != nil {
				b.Logger.Error("Error while playing song: ", err)
			}
			return
		}
		if _, err := b.Client.Rest().UpdateInteractionResponse(i.ApplicationID(), i.Token(), responses.UpdateSuccessComponentsf(p, "modules.music.commands.play.now.playing", []any{track.Info().Title, *track.Info().URI}, getMusicControllerComponents(track))); err != nil {
			b.Logger.Error("Error while updating interaction message: ", err)
		}
	} else {
		if _, err := b.Client.Rest().UpdateInteractionResponse(i.ApplicationID(), i.Token(), responses.UpdateErrorComponentsf(p, "modules.music.commands.play.added.to.queue", []any{len(tracks)}, getMusicControllerComponents(nil))); err != nil {
			b.Logger.Error("Error while updating interaction message: ", err)
		}
	}
	if len(tracks) > 0 {
		player.Queue.Push(tracks...)
	}
}

func giveSearchSelection(b *kbot.Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent, tracks []lavalink.AudioTrack) {
	var options []discord.SelectMenuOption
	for i, track := range tracks {
		if len(options) >= 25 {
			break
		}
		label := track.Info().Title
		if len(label) > 80 {
			label = label[:79] + "…"
		}
		description := "by: " + track.Info().Author
		if len(description) > 100 {
			description = description[:99] + "…"
		}

		options = append(options, discord.SelectMenuOption{
			Label:       label,
			Description: description,
			Value:       strconv.Itoa(i),
		})
	}
	if _, err := e.Client().Rest().UpdateInteractionResponse(e.ApplicationID(), e.Token(), responses.UpdateErrorComponentsf(p, "modules.music.commands.play.search.selection", []any{len(tracks)})); err != nil {
		b.Logger.Error("Error while updating interaction message: ", err)
	}

	if _, err := e.Client().Rest().UpdateInteractionResponse(e.ApplicationID(), e.Token(),
		responses.UpdateSuccessComponentsf(p, "modules.music.autocomplete.select.songs", nil, discord.NewActionRow(
			discord.NewSelectMenu(discord.CustomID("play:search:"+e.ID()), p.Sprintf("modules.music.commands.play.select.songs"), options...).WithMaxValues(len(options)),
		)),
	); err != nil {
		b.Logger.Error("Error while updating interaction message: ", err)
	}

	go func() {
		collectorChan, cancel := bot.NewEventCollector(e.Client(), func(e *events.ComponentInteractionEvent) bool {
			return e.Data.CustomID() == discord.CustomID("play:search:"+e.ID())
		})
		defer cancel()
		for {
			select {
			case e := <-collectorChan:
				if voiceState, ok := e.Client().Caches().VoiceStates().Get(*e.GuildID(), e.User().ID); !ok || voiceState.ChannelID == nil {
					if err := e.CreateMessage(responses.CreateErrorf(p, "modules.music.not.in.voice")); err != nil {
						b.Logger.Error("Failed to update error message: ", err)
					}
					continue
				}
				if err := e.DeferUpdateMessage(); err != nil {
					b.Logger.Error(err)
					return
				}
				var playTracks []lavalink.AudioTrack
				for _, value := range e.SelectMenuInteractionData().Values {
					index, _ := strconv.Atoi(value)
					playTracks = append(playTracks, tracks[index])
				}
				playAndQueue(b, p, e.BaseInteraction, playTracks...)
				return

			case <-time.After(time.Second * 30):
				if _, err := e.Client().Rest().UpdateInteractionResponse(e.ApplicationID(), e.Token(), responses.UpdateErrorComponentsf(p, "modules.music.commands.play.search.timed.out", nil)); err != nil {
					b.Logger.Error("Error while updating interaction message: ", err)
				}
				return
			}
		}
	}()
}

func queueHandler(b *kbot.Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) error {
	tracks := b.MusicPlayers.Get(*e.GuildID()).Queue.Tracks()

	var (
		pages         []string
		page          string
		tracksCounter int
	)
	for i, track := range tracks {
		trackStr := fmt.Sprintf("%d. [`%s`](<%s>) - %s [<@%s>]\n", i+1, track.Info().Title, *track.Info().URI, track.Info().Length, track.UserData().(kbot.AudioTrackData).Requester)
		if len(page)+len(trackStr) > 4096 || tracksCounter >= 10 {
			pages = append(pages, page)
			page = ""
			tracksCounter = 0
		}
		page += trackStr
		tracksCounter++
	}
	if len(page) > 0 {
		pages = append(pages, page)
	}

	return b.Paginator.Create(e.Respond, &paginator.Paginator{
		PageFunc: func(page int, embed *discord.EmbedBuilder) {
			embed.SetTitlef(p.Sprintf("modules.music.commands.queue.title", len(tracks))).SetDescription(pages[page])
		},
		MaxPages:        len(pages),
		ExpiryLastUsage: true,
		ID:              e.ID().String(),
	})
}

func historyHandler(b *kbot.Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) error {
	tracks := b.MusicPlayers.Get(*e.GuildID()).History.Tracks()

	var (
		pages         []string
		page          string
		tracksCounter int
	)
	for i := len(tracks) - 1; i >= 0; i-- {
		track := tracks[i]
		trackStr := fmt.Sprintf("%d. [`%s`](<%s>) - %s [<@%s>]\n", len(tracks)-i, track.Info().Title, *track.Info().URI, track.Info().Length, track.UserData().(kbot.AudioTrackData).Requester)
		if len(page)+len(trackStr) > 4096 || tracksCounter >= 10 {
			pages = append(pages, page)
			page = ""
			tracksCounter = 0
		}
		page += trackStr
		tracksCounter++
	}
	if len(page) > 0 {
		pages = append(pages, page)
	}

	return b.Paginator.Create(e.Respond, &paginator.Paginator{
		PageFunc: func(page int, embed *discord.EmbedBuilder) {
			embed.SetTitlef(p.Sprintf("modules.music.commands.history.title", len(tracks))).SetDescription(pages[page])
		},
		MaxPages:        len(pages),
		ExpiryLastUsage: true,
		ID:              e.ID().String(),
	})
}

func removeSongHandler(b *kbot.Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) error {
	player := b.MusicPlayers.Get(*e.GuildID())
	strIndex := e.SlashCommandInteractionData().String("song")
	index, err := strconv.Atoi(strIndex)
	if err != nil {
		return e.CreateMessage(discord.MessageCreate{Content: p.Sprintf("modules.music.commands.remove.invalid.index")})
	}

	removeTrack := player.Queue.Get(index - 1)
	if removeTrack == nil {
		return e.CreateMessage(discord.MessageCreate{Content: p.Sprintf("modules.music.commands.remove.track.not.found", index)})
	}

	player.Queue.Remove(index - 1)
	return e.CreateMessage(discord.MessageCreate{
		Content: p.Sprintf("modules.music.commands.remove.removed", removeTrack.Info().Title, *removeTrack.Info().URI, index),
	})
}

func removeUserSongsHandler(b *kbot.Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) error {
	player := b.MusicPlayers.Get(*e.GuildID())
	userID := e.SlashCommandInteractionData().Snowflake("user")

	removedTracks := 0
	for i, track := range player.Queue.Tracks() {
		if track.UserData().(kbot.AudioTrackData).Requester == userID {
			player.Queue.Remove(i - removedTracks)
			removedTracks++
		}
	}
	var msg string
	if removedTracks == 0 {
		msg = p.Sprintf("modules.music.commands.remove.no.user.tracks", userID)
	} else {
		msg = p.Sprintf("modules.music.commands.remove.removed.user.tracks", removedTracks, userID)
	}

	return e.CreateMessage(discord.NewMessageCreateBuilder().
		SetContent(msg).
		SetAllowedMentions(&discord.AllowedMentions{}).
		Build(),
	)
}

func clearQueueHandler(b *kbot.Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) error {
	b.MusicPlayers.Get(*e.GuildID()).Queue.Clear()
	return e.CreateMessage(discord.MessageCreate{Content: p.Sprintf("modules.music.commands.clear.cleared")})
}

func stopHandler(b *kbot.Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) error {
	player := b.MusicPlayers.Get(*e.GuildID())
	if err := player.Destroy(); err != nil {
		b.Logger.Error("Failed to destroy player: ", err)
		err = e.CreateMessage(discord.MessageCreate{Content: p.Sprintf("modules.music.commands.stop.error")})
		if err != nil {
			b.Logger.Error("Failed to send message: ", err)
		}
		return err
	}
	if err := b.Client.Disconnect(context.TODO(), *e.GuildID()); err != nil {
		b.Logger.Error("Failed to disconnect kbot: ", err)
		err = e.CreateMessage(discord.MessageCreate{Content: p.Sprintf("modules.music.commands.stop.disconnect.error")})
		if err != nil {
			b.Logger.Error("Failed to send message: ", err)
		}
		return err
	}
	b.MusicPlayers.Delete(*e.GuildID())
	return e.CreateMessage(discord.MessageCreate{Content: p.Sprintf("modules.music.commands.stop.stopped")})
}

func loopHandler(b *kbot.Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) error {
	data := e.SlashCommandInteractionData()
	player := b.MusicPlayers.Get(*e.GuildID())
	loopingType := kbot.LoopingType(data.Int("looping-type"))
	player.Queue.SetType(loopingType)
	emoji := ""
	switch loopingType {
	case kbot.LoopingTypeRepeatSong:
		emoji = "🔂"
	case kbot.LoopingTypeRepeatQueue:
		emoji = "🔁"
	}
	return e.CreateMessage(discord.MessageCreate{
		Content: p.Sprintf("modules.commands.loop", emoji, loopingType),
	})
}

func nowPlayingHandler(b *kbot.Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) error {
	player := b.MusicPlayers.Get(*e.GuildID())

	track := player.PlayingTrack()

	if track == nil {
		return e.CreateMessage(responses.CreateErrorf(p, "modules.music.commands.nowplaying.no.track"))
	}
	i := track.Info()
	embed := discord.NewEmbedBuilder().
		SetAuthorName(p.Sprintf("modules.music.commands.nowplaying.title")).
		SetTitle(i.Title).
		SetURL(*i.URI).
		AddField(p.Sprintf("modules.music.commands.nowplaying.author"), i.Author, true).
		AddField(p.Sprintf("modules.music.commands.nowplaying.requested.by"), discord.UserMention(track.UserData().(kbot.AudioTrackData).Requester), true).
		AddField(p.Sprintf("modules.music.commands.nowplaying.volume"), fmt.Sprintf("%d%%", player.Volume()), true).
		SetThumbnail(getArtworkURL(player.PlayingTrack())).
		SetFooterText(p.Sprintf("modules.music.commands.nowplaying.footer", player.Queue.Len()))
	if !i.IsStream {
		bar := [10]string{"▬", "▬", "▬", "▬", "▬", "▬", "▬", "▬", "▬", "▬"}
		t1 := player.Position()
		t2 := i.Length
		p := int(float64(t1) / float64(t2) * 10)
		bar[p] = "🔘"
		loopString := ""
		if player.Queue.LoopingType() == kbot.LoopingTypeRepeatSong {
			loopString = "🔂"
		}
		if player.Queue.LoopingType() == kbot.LoopingTypeRepeatQueue {
			loopString = "🔁"
		}
		embed.Description += fmt.Sprintf("\n\n%s / %s %s\n%s", formatPosition(t1), formatPosition(t2), loopString, bar)
	}
	return e.CreateMessage(responses.CreateSuccessEmbedComponents(embed.Build(), getMusicControllerComponents(track)))
}

func pauseHandler(b *kbot.Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) error {
	player := b.MusicPlayers.Get(*e.GuildID())
	pause := !player.Paused()
	if err := player.Pause(pause); err != nil {
		var msg string
		if pause {
			msg = p.Sprintf("modules.music.commands.pause.error")
		} else {
			msg = p.Sprintf("modules.music.commands.unpause.error")
		}
		return e.CreateMessage(responses.CreateSuccessf(p, msg))
	}
	var msg string
	if pause {
		msg = p.Sprintf("modules.music.commands.pause")
	} else {
		msg = p.Sprintf("modules.music.commands.unpause")
	}
	return e.CreateMessage(responses.CreateSuccessComponentsf(p, msg, nil, getMusicControllerComponents(player.PlayingTrack())))
}

func volumeHandler(b *kbot.Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) error {
	player := b.MusicPlayers.Get(*e.GuildID())
	volume := e.SlashCommandInteractionData().Int("volume")
	if err := player.SetVolume(volume); err != nil {
		return e.CreateMessage(responses.CreateErrorf(p, "modules.music.commands.volume.error"))
	}
	return e.CreateMessage(responses.CreateSuccessf(p, "modules.music.commands.volume.success", volume))
}

func bassBoostHandler(b *kbot.Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) error {
	player := b.MusicPlayers.Get(*e.GuildID())
	enable := e.SlashCommandInteractionData().Bool("enable")

	if enable {
		if err := player.Filters().SetEqualizer(bassBoost).Commit(); err != nil {
			return e.CreateMessage(responses.CreateErrorf(p, "modules.music.commands.bass.boost.enable.error"))
		}
		return e.CreateMessage(responses.CreateSuccessf(p, "modules.music.commands.bass.boost.enable.success"))
	}
	if err := player.Filters().SetEqualizer(&lavalink.Equalizer{}).Commit(); err != nil {
		return e.CreateMessage(responses.CreateErrorf(p, "modules.music.commands.bass.boost.disable.error"))
	}
	return e.CreateMessage(responses.CreateSuccessf(p, "modules.music.commands.bass.boost.disable.success"))
}

func seekHandler(b *kbot.Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) error {
	data := e.SlashCommandInteractionData()
	player := b.MusicPlayers.Get(*e.GuildID())
	position := data.Int("position")
	timeUnit := lavalink.Second
	if timeUnitPtr, ok := data.OptInt("time-unit"); ok {
		timeUnit = lavalink.Duration(timeUnitPtr)
	}

	finalPosition := lavalink.Duration(position) * timeUnit
	if finalPosition > player.PlayingTrack().Info().Length {
		return e.CreateMessage(responses.CreateErrorf(p, "modules.music.commands.seek.position.too.big"))
	}
	if err := player.Seek(finalPosition); err != nil {
		return e.CreateMessage(responses.CreateErrorf(p, "modules.music.commands.seek.error"))
	}
	return e.CreateMessage(responses.CreateSuccessf(p, "modules.music.commands.seek.success"))
}

func nextHandler(b *kbot.Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) error {
	player := b.MusicPlayers.Get(*e.GuildID())
	nextTrack := player.Queue.Pop()

	if nextTrack == nil {
		return e.CreateMessage(responses.CreateErrorf(p, "modules.music.commands.next.no.track"))
	}

	if err := player.Play(nextTrack); err != nil {
		return e.CreateMessage(responses.CreateErrorf(p, "modules.music.commands.next.error"))
	}
	return e.CreateMessage(responses.CreateSuccessComponentsf(p, "modules.music.commands.next.success", []any{nextTrack.Info().Title, *nextTrack.Info().URI, nextTrack.Info().Length}, getMusicControllerComponents(nextTrack)))
}

func previousHandler(b *kbot.Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) error {
	player := b.MusicPlayers.Get(*e.GuildID())
	previousTrack := player.History.Last()

	if previousTrack == nil {
		return e.CreateMessage(responses.CreateErrorf(p, "modules.music.commands.previous.no.track"))
	}

	if err := player.Play(previousTrack); err != nil {
		return e.CreateMessage(responses.CreateErrorf(p, "modules.music.commands.previous.error"))
	}
	return e.CreateMessage(responses.CreateSuccessComponentsf(p, "modules.music.commands.previous.success", []any{previousTrack.Info().Title, *previousTrack.Info().URI, previousTrack.Info().Length}, getMusicControllerComponents(previousTrack)))
}

func shuffleHandler(b *kbot.Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) error {
	queue := b.MusicPlayers.Get(*e.GuildID()).Queue

	if queue.Len() == 0 {
		return e.CreateMessage(responses.CreateErrorf(p, "modules.music.commands.shuffle.no.track"))
	}
	return e.CreateMessage(responses.CreateSuccessf(p, "modules.music.commands.shuffle.success"))
}

func formatPosition(position lavalink.Duration) string {
	if position == 0 {
		return "0:00"
	}
	return fmt.Sprintf("%d:%02d", position.Minutes(), position.SecondsPart())
}

func getArtworkURL(track lavalink.AudioTrack) string {
	switch track.Info().SourceName {
	case "youtube":
		return "https://i.ytimg.com/vi/" + track.Info().Identifier + "/maxresdefault.jpg"
	case "twitch":
		return "https://static-cdn.jtvnw.net/previews-ttv/live_user_" + track.Info().Identifier + "-440x248.jpg"

	case "spotify", "applemusic":
		if isrcTrack, ok := track.(*source_extensions.ISRCAudioTrack); ok && isrcTrack.ArtworkURL != nil {
			return *isrcTrack.ArtworkURL
		}
	}
	return ""
}

func likedSongsListHandler(b *kbot.Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) error {
	tracks, err := b.DB.LikedSongs().GetAll(e.User().ID)
	if err != nil {
		return err
	}
	if len(tracks) == 0 {
		return e.CreateMessage(responses.CreateErrorf(p, "modules.music.commands.liked.songs.list.empty"))
	}
	var (
		pages         []string
		page          string
		tracksCounter int
	)
	for i, track := range tracks {
		trackStr := fmt.Sprintf("%d. [`%s`](<%s>)\n", i+1, track.Title, track.Query)
		if len(page)+len(trackStr) > 4096 || tracksCounter >= 10 {
			pages = append(pages, page)
			page = ""
			tracksCounter = 0
		}
		page += trackStr
		tracksCounter++
	}
	if len(page) > 0 {
		pages = append(pages, page)
	}

	return b.Paginator.Create(e.Respond, &paginator.Paginator{
		PageFunc: func(page int, embed *discord.EmbedBuilder) {
			embed.SetTitlef(p.Sprintf("modules.music.commands.liked.songs.list.title", len(tracks))).SetDescription(pages[page])
		},
		MaxPages:        len(pages),
		ExpiryLastUsage: true,
		ID:              e.ID().String(),
	})
}

func likedSongsRemoveHandler(b *kbot.Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) error {
	songName := e.SlashCommandInteractionData().String("song")

	if err := b.DB.LikedSongs().Delete(e.User().ID, songName); err != nil {
		return e.CreateMessage(responses.CreateErrorf(p, "modules.music.commands.liked.songs.remove.error"))
	}
	return e.CreateMessage(responses.CreateSuccessf(p, "modules.music.commands.liked.songs.remove.success", songName))
}

func likedSongsClearHandler(b *kbot.Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) error {
	if err := b.DB.LikedSongs().DeleteAll(e.User().ID); err != nil {
		return e.CreateMessage(responses.CreateErrorf(p, "modules.music.commands.liked.songs.clear.error"))
	}
	return e.CreateMessage(responses.CreateSuccessf(p, "modules.music.commands.liked.songs.clear.success"))
}

func likedSongsPlayHandler(b *kbot.Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) error {
	return nil
}
