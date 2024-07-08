package commands

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/disgolink/v3/disgolink"
	"github.com/disgoorg/disgolink/v3/lavalink"
	"github.com/disgoorg/lavaqueue-plugin"
	"github.com/disgoorg/snowflake/v2"
	"github.com/topi314/tint"

	"github.com/KittyBot-Org/KittyBotGo/service/bot/res"
)

var (
	urlPattern    = regexp.MustCompile("^https?://[-a-zA-Z0-9+&@#/%?=~_|!:,.;]*[-a-zA-Z0-9+&@#/%=~_|]?")
	searchPattern = regexp.MustCompile(`^(.{2})(search|isrc):(.+)`)
	queryTypes    = []string{"liked_track", "playlist", "play_history"}
)

func (c *commands) OnPlayerPlay(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	query := data.String("query")

	var (
		id       int
		loadType string
	)
	parts := strings.SplitN(query, ":", 2)
	if len(parts) == 2 && slices.Contains(queryTypes, parts[0]) {
		if loadID, err := strconv.Atoi(parts[1]); err == nil {
			id = loadID
			loadType = parts[0]
		}
	} else if !urlPattern.MatchString(query) && !searchPattern.MatchString(query) {
		if source, ok := data.OptString("source"); ok {
			query = lavalink.SearchType(source).Apply(query)
		} else {
			query = lavalink.SearchTypeYouTube.Apply(query)
		}
	}

	voiceState, ok := c.Discord.Caches().VoiceState(*e.GuildID(), e.User().ID)
	if !ok {
		return e.CreateMessage(res.CreateError("You are not in a voice channel"))
	}

	player := c.Lavalink.Player(*e.GuildID())

	if err := e.DeferCreateMessage(false); err != nil {
		return err
	}

	switch loadType {
	case "liked_track":
		track, err := c.Database.GetLikedTrack(id)
		if err != nil {
			_, err = e.UpdateInteractionResponse(res.UpdateErr("Failed to get liked song", err))
			return err
		}
		return c.handleTracks(e, *voiceState.ChannelID, track.Track)

	case "playlist":
		_, playlistTracks, err := c.Database.GetPlaylist(id)
		if err != nil {
			_, err = e.UpdateInteractionResponse(res.UpdateErr("Failed to get playlist", err))
			return err
		}

		tracks := make([]lavalink.Track, len(playlistTracks))
		for i, track := range playlistTracks {
			tracks[i] = track.Track
		}
		return c.handleTracks(e, *voiceState.ChannelID, tracks...)

	case "play_history":
		track, err := c.Database.GetPlayHistoryTrack(id)
		if err != nil {
			_, err = e.UpdateInteractionResponse(res.UpdateErr("Failed to get play history song", err))
			return err
		}

		return c.handleTracks(e, *voiceState.ChannelID, track.Track)
	}

	var loadErr error
	player.Node().LoadTracksHandler(e.Ctx, query, disgolink.NewResultHandler(
		func(track lavalink.Track) {
			loadErr = c.handleTracks(e, *voiceState.ChannelID, track)
		},
		func(playlist lavalink.Playlist) {
			loadErr = c.handleTracks(e, *voiceState.ChannelID, playlist.Tracks...)
		},
		func(tracks []lavalink.Track) {
			loadErr = c.handleTracks(e, *voiceState.ChannelID, tracks[0])
		},
		func() {
			_, loadErr = e.UpdateInteractionResponse(res.UpdateError("No results found for %s", query))
		},
		func(err error) {
			_, loadErr = e.UpdateInteractionResponse(res.UpdateErr("An error occurred", err))
		},
	))

	return loadErr
}

func (c *commands) handleTracks(e *handler.CommandEvent, channelID snowflake.ID, tracks ...lavalink.Track) error {
	_, ok := c.Discord.Caches().VoiceState(*e.GuildID(), e.ApplicationID())
	if !ok {
		if err := c.Discord.UpdateVoiceState(context.Background(), *e.GuildID(), &channelID, false, false); err != nil {
			_, err = e.UpdateInteractionResponse(res.UpdateErr("An error occurred", err))
			return err
		}
	}

	queueTracks := make([]lavaqueue.QueueTrack, len(tracks))
	for i, track := range tracks {
		queueTracks[i] = lavaqueue.QueueTrack{
			Encoded:  track.Encoded,
			UserData: nil, // TODO: Add user data
		}
	}

	player := c.Lavalink.Player(*e.GuildID())
	track, err := lavaqueue.AddQueueTracks(e.Ctx, player.Node(), *e.GuildID(), queueTracks)
	if err != nil {
		_, err = e.UpdateInteractionResponse(res.UpdateErr("An error occurred playing the song", err))
		return err
	}

	var (
		content     string
		likeButton  bool
		tracksCount = len(tracks)
	)
	if track != nil {
		content = fmt.Sprintf("▶ Playing: %s", res.FormatTrack(*track, 0))
		likeButton = true
		tracksCount--
	}
	if len(tracks) > 0 {
		content += fmt.Sprintf("\nAdded %d songs to the queue", tracksCount)
	}

	go func() {
		if err := c.Database.AddPlayHistoryTracks(e.User().ID, tracks); err != nil {
			slog.Error("error adding play history songs", tint.Err(err))
		}
	}()

	_, err = e.UpdateInteractionResponse(res.UpdatePlayer(content, likeButton))
	return err
}

func (c *commands) OnPlayerPlayAutocomplete(e *handler.AutocompleteEvent) error {
	query := e.Data.String("query")

	limit := 24
	if strings.TrimSpace(query) == "" {
		limit = 25
	}

	tracks, err := c.Database.SearchPlay(e.User().ID, query, limit)
	if err != nil {
		slog.Error("error searching play", tint.Err(err))
		return e.AutocompleteResult(nil)
	}

	choices := make([]discord.AutocompleteChoice, 0, 25)
	if limit == 24 {
		choices = append(choices, discord.AutocompleteChoiceString{
			Name:  res.Trim(fmt.Sprintf("🔎 %s", query), 100),
			Value: query,
		})
	}

	for _, track := range tracks {
		var prefix string
		switch track.Type {
		case "liked_track":
			prefix = "❤ "
		case "play_history":
			prefix = "🕒 "
		case "playlist":
			prefix = "📜 "
		}

		choices = append(choices, discord.AutocompleteChoiceString{
			Name:  res.Trim(prefix+track.Name, 100),
			Value: fmt.Sprintf("%s:%d", track.Type, track.ID),
		})
	}

	return e.AutocompleteResult(choices)
}
