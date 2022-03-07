package music

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/DisgoOrg/disgo/core"
	"github.com/DisgoOrg/disgo/core/events"
	"github.com/DisgoOrg/disgo/discord"
	"github.com/DisgoOrg/disgolink/lavalink"
	"github.com/DisgoOrg/snowflake"
	"github.com/DisgoOrg/source-extensions-plugin"
	"github.com/DisgoOrg/utils/paginator"
	"github.com/KittyBot-Org/KittyBotGo/internal/models"
	"github.com/KittyBot-Org/KittyBotGo/internal/types"
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

func playHandler(b *types.Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) error {
	data := e.SlashCommandInteractionData()

	query := *data.Options.String("query")
	if searchProvider := data.Options.String("searchProvider"); searchProvider != nil {
		query = lavalink.SearchType(*searchProvider).Apply(query)
	} else {
		if !urlPattern.MatchString(query) && !searchPattern.MatchString(query) {
			query = lavalink.SearchTypeYoutube.Apply(query)
		}
	}

	if err := e.DeferCreateMessage(false); err != nil {
		return err
	}
	return b.Lavalink.BestRestClient().LoadItemHandler(context.TODO(), query, lavalink.NewResultHandler(
		func(track lavalink.AudioTrack) {
			b.AddPlayHistory(e.User.ID, track.Info().Title, query)
			playAndQueue(b, p, e.CreateInteraction, track)
		},
		func(playlist lavalink.AudioPlaylist) {
			b.AddPlayHistory(e.User.ID, playlist.Name(), query)
			playAndQueue(b, p, e.CreateInteraction, playlist.Tracks()...)
		},
		func(tracks []lavalink.AudioTrack) {
			b.AddPlayHistory(e.User.ID, *data.Options.String("query"), query)
			giveSearchSelection(b, p, e, tracks)
		},
		func() {
			if _, err := e.UpdateOriginalMessage(discord.NewMessageUpdateBuilder().SetContent(p.Sprintf("modules.music.commands.play.no.results")).Build()); err != nil {
				b.Logger.Error(err)
			}
		},
		func(ex lavalink.FriendlyException) {
			if _, err := e.UpdateOriginalMessage(discord.NewMessageUpdateBuilder().SetContent(p.Sprintf("modules.music.commands.play.exception", ex.Message)).Build()); err != nil {
				b.Logger.Error(err)
			}
		},
	))
}

func playAndQueue(b *types.Bot, p *message.Printer, i core.CreateInteraction, tracks ...lavalink.AudioTrack) {
	player := b.MusicPlayers.Get(*i.GuildID)
	if player == nil {
		player = b.MusicPlayers.New(*i.GuildID, types.PlayerTypeMusic, types.LoopingTypeOff)
		b.MusicPlayers.Add(player)
	}
	var voiceChannelID snowflake.Snowflake
	if voiceState := i.Member.VoiceState(); voiceState == nil || voiceState.ChannelID == nil {
		if _, err := i.UpdateOriginalMessage(discord.NewMessageUpdateBuilder().SetContent(p.Sprintf("modules.music.not.in.voice")).ClearContainerComponents().Build()); err != nil {
			b.Logger.Error(err)
		}
		return
	} else {
		voiceChannelID = *voiceState.ChannelID
	}
	if voiceState := i.Guild().SelfMember().VoiceState(); voiceState == nil || voiceState.ChannelID == nil || *voiceState.ChannelID != voiceChannelID {
		if err := b.Bot.AudioController.Connect(context.TODO(), *i.GuildID, voiceChannelID); err != nil {
			if _, err = i.UpdateOriginalMessage(discord.NewMessageUpdateBuilder().SetContent(p.Sprintf("modules.music.no.permissions")).ClearContainerComponents().Build()); err != nil {
				b.Logger.Error(err)
			}
			return
		}
	}

	for ii := range tracks {
		tracks[ii].SetUserData(models.AudioTrackData{
			Requester: i.User.ID,
		})
	}

	if player.PlayingTrack() == nil {
		track := tracks[0]
		if len(tracks) > 0 {
			tracks = tracks[1:]
		}
		if err := player.Play(track); err != nil {
			if _, err = i.UpdateOriginalMessage(discord.NewMessageUpdateBuilder().SetContent(p.Sprintf("modules.music.commands.play.error")).ClearContainerComponents().Build()); err != nil {
				b.Logger.Error("Error while playing song: ", err)
			}
			return
		}
		if _, err := i.UpdateOriginalMessage(discord.NewMessageUpdateBuilder().
			SetContent(p.Sprintf("modules.music.commands.play.now.playing", track.Info().Title, *track.Info().URI)).
			SetContainerComponents(getMusicControllerComponents(track)).
			Build(),
		); err != nil {
			b.Logger.Error("Error while updating interaction message: ", err)
		}
	} else {
		if _, err := i.UpdateOriginalMessage(discord.NewMessageUpdateBuilder().
			SetContent(p.Sprintf("modules.music.commands.play.added.to.queue", len(tracks))).
			SetContainerComponents(getMusicControllerComponents(nil)).
			Build(),
		); err != nil {
			b.Logger.Error("Error while updating interaction message: ", err)
		}
	}
	if len(tracks) > 0 {
		player.Queue.Push(tracks...)
	}
}

func giveSearchSelection(b *types.Bot, p *message.Printer, event *events.ApplicationCommandInteractionEvent, tracks []lavalink.AudioTrack) {
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
	if _, err := event.UpdateOriginalMessage(discord.NewMessageUpdateBuilder().
		SetContent(p.Sprintf("modules.music.autocomplete.select.songs")).
		AddActionRow(discord.NewSelectMenu(discord.CustomID("play:search:"+event.ID), p.Sprintf("modules.music.commands.play.select.songs"), options...).WithMaxValues(len(options))).
		Build()); err != nil {
		b.Logger.Error(err)
	}
	go func() {
		coll, cancel := b.Bot.Collectors.NewComponentInteractionCollector(func(interaction *core.ComponentInteraction) bool {
			return interaction.Data.ID() == discord.CustomID("play:search:"+event.ID)
		})
		defer cancel()
		for {
			select {
			case i := <-coll:
				if voiceState := i.Member.VoiceState(); voiceState == nil || voiceState.ChannelID == nil {
					if err := i.CreateMessage(discord.NewMessageCreateBuilder().SetContent(p.Sprintf("modules.music.not.in.voice")).SetEphemeral(true).Build()); err != nil {
						b.Logger.Error(err)
					}
					continue
				}
				if err := i.DeferUpdateMessage(); err != nil {
					b.Logger.Error(err)
					return
				}
				var playTracks []lavalink.AudioTrack
				for _, value := range i.SelectMenuInteractionData().Values {
					index, _ := strconv.Atoi(value)
					playTracks = append(playTracks, tracks[index])
				}
				playAndQueue(b, p, i.CreateInteraction, playTracks...)
				return

			case <-time.After(time.Second * 30):
				if _, err := event.UpdateOriginalMessage(discord.NewMessageUpdateBuilder().SetContent(p.Sprintf("modules.music.commands.play.search.timed.out")).ClearContainerComponents().Build()); err != nil {
					b.Logger.Error(err)
				}
				return
			}
		}
	}()
}

func queueHandler(b *types.Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) error {
	tracks := b.MusicPlayers.Get(*e.GuildID).Queue.Tracks()

	var (
		pages         []string
		page          string
		tracksCounter int
	)
	for i, track := range tracks {
		trackStr := fmt.Sprintf("%d. [`%s`](<%s>) - %s [<@%s>]\n", i+1, track.Info().Title, *track.Info().URI, track.Info().Length, track.UserData().(models.AudioTrackData).Requester)
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

	return b.Paginator.Create(e.CreateInteraction, &paginator.Paginator{
		PageFunc: func(page int, embed *discord.EmbedBuilder) {
			embed.SetTitlef(p.Sprintf("modules.music.commands.queue.title", len(tracks))).SetDescription(pages[page])
		},
		MaxPages:        len(pages),
		ExpiryLastUsage: true,
	})
}

func historyHandler(b *types.Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) error {
	tracks := b.MusicPlayers.Get(*e.GuildID).History.Tracks()

	var (
		pages         []string
		page          string
		tracksCounter int
	)
	for i := len(tracks) - 1; i >= 0; i-- {
		track := tracks[i]
		trackStr := fmt.Sprintf("%d. [`%s`](<%s>) - %s [<@%s>]\n", len(tracks)-i, track.Info().Title, *track.Info().URI, track.Info().Length, track.UserData().(models.AudioTrackData).Requester)
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

	return b.Paginator.Create(e.CreateInteraction, &paginator.Paginator{
		PageFunc: func(page int, embed *discord.EmbedBuilder) {
			embed.SetTitlef(p.Sprintf("modules.music.commands.history.title", len(tracks))).SetDescription(pages[page])
		},
		MaxPages:        len(pages),
		ExpiryLastUsage: true,
	})
}

func removeSongHandler(b *types.Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) error {
	player := b.MusicPlayers.Get(*e.GuildID)
	strIndex := *e.SlashCommandInteractionData().Options.String("song")
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

func removeUserSongsHandler(b *types.Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) error {
	player := b.MusicPlayers.Get(*e.GuildID)
	userID := *e.SlashCommandInteractionData().Options.Snowflake("user")

	removedTracks := 0
	for i, track := range player.Queue.Tracks() {
		if track.UserData().(models.AudioTrackData).Requester == userID {
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

func clearQueueHandler(b *types.Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) error {
	b.MusicPlayers.Get(*e.GuildID).Queue.Clear()
	return e.CreateMessage(discord.MessageCreate{Content: p.Sprintf("modules.music.commands.clear.cleared")})
}

func stopHandler(b *types.Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) error {
	player := b.MusicPlayers.Get(*e.GuildID)
	if err := player.Destroy(); err != nil {
		b.Logger.Error("Failed to destroy player: ", err)
		err = e.CreateMessage(discord.MessageCreate{Content: p.Sprintf("modules.music.commands.stop.error")})
		if err != nil {
			b.Logger.Error("Failed to send message: ", err)
		}
		return err
	}
	if err := b.Bot.AudioController.Disconnect(context.TODO(), *e.GuildID); err != nil {
		b.Logger.Error("Failed to disconnect bot: ", err)
		err = e.CreateMessage(discord.MessageCreate{Content: p.Sprintf("modules.music.commands.stop.disconnect.error")})
		if err != nil {
			b.Logger.Error("Failed to send message: ", err)
		}
		return err
	}
	b.MusicPlayers.Delete(*e.GuildID)
	return e.CreateMessage(discord.MessageCreate{Content: p.Sprintf("modules.music.commands.stop.stopped")})
}

func loopHandler(b *types.Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) error {
	data := e.SlashCommandInteractionData()
	player := b.MusicPlayers.Get(*e.GuildID)
	loopingType := types.LoopingType(*data.Options.Int("looping-type"))
	player.Queue.SetType(loopingType)
	emoji := ""
	switch loopingType {
	case types.LoopingTypeRepeatSong:
		emoji = "🔂"
	case types.LoopingTypeRepeatQueue:
		emoji = "🔁"
	}
	return e.CreateMessage(discord.MessageCreate{
		Content: p.Sprintf("modules.commands.loop", emoji, loopingType),
	})
}

func nowPlayingHandler(b *types.Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) error {
	player := b.MusicPlayers.Get(*e.GuildID)

	track := player.PlayingTrack()
	i := track.Info()
	embed := discord.NewEmbedBuilder().
		SetColor(types.KittyBotColor).
		SetAuthorName(p.Sprintf("modules.music.commands.now.playing.title")).
		SetTitle(i.Title).
		SetURL(*i.URI).
		AddField(p.Sprintf("modules.music.commands.now.playing.author"), i.Author, true).
		AddField(p.Sprintf("modules.music.commands.now.playing.requested.by"), fmt.Sprintf("<@%s>", track.UserData().(models.AudioTrackData).Requester), true).
		AddField(p.Sprintf("modules.music.commands.now.playing.volume"), fmt.Sprintf("%d%%", player.Volume()), true).
		SetThumbnail(getArtworkURL(player.PlayingTrack())).
		SetFooterText(p.Sprintf("modules.music.commands.now.playing.footer", player.Queue.Len()))
	if !i.IsStream {
		bar := [10]string{"▬", "▬", "▬", "▬", "▬", "▬", "▬", "▬", "▬", "▬"}
		t1 := player.Position()
		t2 := i.Length
		p := int(float64(t1) / float64(t2) * 10)
		bar[p] = "🔘"
		loopString := ""
		if player.Queue.LoopingType() == types.LoopingTypeRepeatSong {
			loopString = "🔂"
		}
		if player.Queue.LoopingType() == types.LoopingTypeRepeatQueue {
			loopString = "🔁"
		}
		embed.Description += fmt.Sprintf("\n\n%s / %s %s\n%s", formatPosition(t1), formatPosition(t2), loopString, bar)
	}
	return e.CreateMessage(discord.MessageCreate{
		Embeds:     []discord.Embed{embed.Build()},
		Components: []discord.ContainerComponent{getMusicControllerComponents(track)},
	})
}

func pauseHandler(b *types.Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) error {
	player := b.MusicPlayers.Get(*e.GuildID)
	pause := !player.Paused()
	if err := player.Pause(pause); err != nil {
		var msg string
		if pause {
			msg = p.Sprintf("modules.music.commands.pause.error")
		} else {
			msg = p.Sprintf("modules.music.commands.unpause.error")
		}
		return e.CreateMessage(discord.MessageCreate{Content: msg, Flags: discord.MessageFlagEphemeral})
	}
	var msg string
	if pause {
		msg = p.Sprintf("modules.music.commands.pause")
	} else {
		msg = p.Sprintf("modules.music.commands.unpause")
	}
	return e.CreateMessage(discord.MessageCreate{
		Content:    msg,
		Components: []discord.ContainerComponent{getMusicControllerComponents(player.PlayingTrack())},
	})
}

func volumeHandler(b *types.Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) error {
	player := b.MusicPlayers.Get(*e.GuildID)
	volume := *e.SlashCommandInteractionData().Options.Int("volume")
	if err := player.SetVolume(volume); err != nil {
		return e.CreateMessage(discord.MessageCreate{Content: p.Sprintf("modules.music.commands.volume.set.error")})
	}
	return e.CreateMessage(discord.MessageCreate{
		Content: p.Sprintf("modules.music.commands.volume.set", volume),
	})
}

func bassBoostHandler(b *types.Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) error {
	player := b.MusicPlayers.Get(*e.GuildID)
	enable := *e.SlashCommandInteractionData().Options.Bool("enable")
	if enable {
		if err := player.Filters().SetEqualizer(bassBoost).Commit(); err != nil {
			return e.CreateMessage(discord.MessageCreate{Content: p.Sprintf("modules.music.commands.bass.boost.enabled.error")})
		}
	} else {
		if err := player.Filters().SetEqualizer(&lavalink.Equalizer{}).Commit(); err != nil {
			return e.CreateMessage(discord.MessageCreate{Content: p.Sprintf("modules.music.commands.bass.boost.disabled.error")})
		}
	}
	var msg string
	if enable {
		msg = p.Sprintf("modules.music.commands.bass.boost.enabled")
	} else {
		msg = p.Sprintf("modules.music.commands.bass.boost.disabled")
	}
	return e.CreateMessage(discord.MessageCreate{
		Content: msg,
	})
}

func seekHandler(b *types.Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) error {
	data := e.SlashCommandInteractionData()
	player := b.MusicPlayers.Get(*e.GuildID)
	position := *data.Options.Int("position")
	timeUnit := lavalink.Second
	if timeUnitPtr := data.Options.Int("time-unit"); timeUnitPtr != nil {
		timeUnit = lavalink.Duration(*timeUnitPtr)
	}

	finalPosition := lavalink.Duration(position) * timeUnit
	if finalPosition > player.PlayingTrack().Info().Length {
		return e.CreateMessage(discord.MessageCreate{Content: p.Sprintf("modules.music.commands.seek.position.too.big"), Flags: discord.MessageFlagEphemeral})
	}
	if err := player.Seek(finalPosition); err != nil {
		return e.CreateMessage(discord.MessageCreate{Content: p.Sprintf("modules.music.commands.seek.error")})
	}
	return e.CreateMessage(discord.MessageCreate{
		Content: p.Sprintf("modules.music.commands.seek.success"),
	})
}

func nextHandler(b *types.Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) error {
	player := b.MusicPlayers.Get(*e.GuildID)
	nextTrack := player.Queue.Pop()

	if err := player.Play(nextTrack); err != nil {
		return e.CreateMessage(discord.MessageCreate{Content: p.Sprintf("modules.music.commands.next.error")})
	}
	return e.CreateMessage(discord.MessageCreate{
		Content:    p.Sprintf("modules.music.commands.next.success", nextTrack.Info().Title, *nextTrack.Info().URI, nextTrack.Info().Length),
		Components: []discord.ContainerComponent{getMusicControllerComponents(nextTrack)},
	})
}

func previousHandler(b *types.Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) error {
	player := b.MusicPlayers.Get(*e.GuildID)
	nextTrack := player.History.Last()

	if err := player.Play(nextTrack); err != nil {
		return e.CreateMessage(discord.MessageCreate{Content: p.Sprintf("modules.music.commands.previous.error")})
	}
	return e.CreateMessage(discord.MessageCreate{
		Content:    p.Sprintf("modules.music.commands.previous.success", nextTrack.Info().Title, *nextTrack.Info().URI, nextTrack.Info().Length),
		Components: []discord.ContainerComponent{getMusicControllerComponents(nextTrack)},
	})
}

func shuffleHandler(b *types.Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) error {
	b.MusicPlayers.Get(*e.GuildID).Queue.Shuffle()
	return e.CreateMessage(discord.MessageCreate{
		Content: p.Sprintf("modules.music.commands.shuffle"),
	})
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
		isrcTrack, ok := track.(*source_extensions.ISRCAudioTrack)
		if ok && isrcTrack.ArtworkURL != nil {
			return *isrcTrack.ArtworkURL
		}
	}
	return ""
}

func likedSongsListHandler(b *types.Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) error {
	var tracks []models.LikedSong
	if err := b.DB.NewSelect().Model(&tracks).Where("user_id = ?", e.User.ID).Scan(context.TODO()); err != nil {
		return err
	}
	if len(tracks) == 0 {
		return e.CreateMessage(discord.MessageCreate{Content: p.Sprintf("modules.music.commands.liked.songs.list.empty")})
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

	return b.Paginator.Create(e.CreateInteraction, &paginator.Paginator{
		PageFunc: func(page int, embed *discord.EmbedBuilder) {
			embed.SetTitlef(p.Sprintf("modules.music.commands.liked.songs.list.title", len(tracks))).SetDescription(pages[page])
		},
		MaxPages:        len(pages),
		ExpiryLastUsage: true,
	})
}

func likedSongsRemoveHandler(b *types.Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) error {
	songName := *e.SlashCommandInteractionData().Options.String("song")

	if _, err := b.DB.NewDelete().Model((*models.LikedSong)(nil)).Where("user_id = ? AND title like ?", e.User.ID, songName).Exec(context.TODO()); err != nil {
		return e.CreateMessage(discord.MessageCreate{Content: p.Sprintf("modules.music.commands.liked.songs.remove.error"), Flags: discord.MessageFlagEphemeral})
	}
	return e.CreateMessage(discord.MessageCreate{Content: p.Sprintf("modules.music.commands.liked.songs.remove.success", songName)})
}

func likedSongsClearHandler(b *types.Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) error {
	if _, err := b.DB.NewDelete().Model((*models.LikedSong)(nil)).Where("user_id = ?", e.User.ID).Exec(context.TODO()); err != nil {
		return e.CreateMessage(discord.MessageCreate{Content: p.Sprintf("modules.music.commands.liked.songs.clear.error"), Flags: discord.MessageFlagEphemeral})
	}
	return e.CreateMessage(discord.MessageCreate{Content: p.Sprintf("modules.music.commands.liked.songs.clear.success")})
}

func likedSongsPlayHandler(b *types.Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) error {
	return nil
}
