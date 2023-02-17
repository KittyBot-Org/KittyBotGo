package handlers

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/disgolink/v2/disgolink"
	"github.com/disgoorg/disgolink/v2/lavalink"
	"github.com/disgoorg/snowflake/v2"
	"golang.org/x/exp/slices"

	"github.com/KittyBot-Org/KittyBotGo/service/bot/res"
)

var (
	urlPattern    = regexp.MustCompile("^https?://[-a-zA-Z0-9+&@#/%?=~_|!:,.;]*[-a-zA-Z0-9+&@#/%=~_|]?")
	searchPattern = regexp.MustCompile(`^(.{2})(search|isrc):(.+)`)
	queryTypes    = []string{"liked_track", "playlist", "play_history"}
)

func (h *Handlers) OnPlayerPlay(e *handler.CommandEvent) error {
	data := e.SlashCommandInteractionData()
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

	voiceState, ok := h.Discord.Caches().VoiceState(*e.GuildID(), e.User().ID)
	if !ok {
		return e.CreateMessage(res.CreateError("You are not in a voice channel"))
	}

	player := h.Lavalink.Player(*e.GuildID())
	if _, err := h.Database.GetPlayer(*e.GuildID(), player.Node().Config().Name); err != nil {
		return e.CreateMessage(res.CreateErr("Failed to get or create player", err))
	}

	if err := e.DeferCreateMessage(false); err != nil {
		return err
	}

	switch loadType {
	case "liked_track":
		track, err := h.Database.GetLikedTrack(id)
		if err != nil {
			_, err = e.UpdateInteractionResponse(res.UpdateErr("Failed to get liked song", err))
			return err
		}
		return h.handleTracks(context.Background(), e, *voiceState.ChannelID, track.Track)

	case "playlist":
		_, playlistTracks, err := h.Database.GetPlaylist(id)
		if err != nil {
			_, err = e.UpdateInteractionResponse(res.UpdateErr("Failed to get playlist", err))
			return err
		}

		tracks := make([]lavalink.Track, len(playlistTracks))
		for i, track := range playlistTracks {
			tracks[i] = track.Track
		}
		return h.handleTracks(context.Background(), e, *voiceState.ChannelID, tracks...)

	case "play_history":
		track, err := h.Database.GetPlayHistoryTrack(id)
		if err != nil {
			_, err = e.UpdateInteractionResponse(res.UpdateErr("Failed to get play history song", err))
			return err
		}

		return h.handleTracks(context.Background(), e, *voiceState.ChannelID, track.Track)
	}

	go func() {
		var loadErr error
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		player.Node().LoadTracks(ctx, query, disgolink.NewResultHandler(
			func(track lavalink.Track) {
				loadErr = h.handleTracks(ctx, e, *voiceState.ChannelID, track)
			},
			func(playlist lavalink.Playlist) {
				loadErr = h.handleTracks(ctx, e, *voiceState.ChannelID, playlist.Tracks...)
			},
			func(tracks []lavalink.Track) {
				loadErr = h.handleTracks(ctx, e, *voiceState.ChannelID, tracks[0])
			},
			func() {
				_, loadErr = e.UpdateInteractionResponse(res.UpdateError("No results found for %s", query))
			},
			func(err error) {
				_, loadErr = e.UpdateInteractionResponse(res.UpdateErr("An error occurred", err))
			},
		))
		if loadErr != nil {
			h.Logger.Errorf("error loading songs: %s", loadErr)
		}
	}()

	return nil
}

func (h *Handlers) handleTracks(ctx context.Context, e *handler.CommandEvent, channelID snowflake.ID, tracks ...lavalink.Track) error {
	_, ok := h.Discord.Caches().VoiceState(*e.GuildID(), e.ApplicationID())
	if !ok {
		if err := h.Discord.UpdateVoiceState(context.Background(), *e.GuildID(), &channelID, false, false); err != nil {
			_, err = e.UpdateInteractionResponse(res.UpdateErr("An error occurred", err))
			return err
		}
	}
	player := h.Lavalink.Player(*e.GuildID())
	var (
		content    string
		likeButton bool
	)
	if player.Track() == nil {
		track := tracks[0]
		tracks = tracks[1:]

		if err := player.Update(ctx, lavalink.WithTrack(track)); err != nil {
			_, err = e.UpdateInteractionResponse(res.UpdateErr("An error occurred", err))
			return err
		}
		content = fmt.Sprintf("▶ Playing: %s", res.FormatTrack(track, 0))
		likeButton = true
	}

	if len(tracks) > 0 {
		content += fmt.Sprintf("\nAdded %d songs to the queue", len(tracks))
		if err := h.Database.AddQueueTracks(*e.GuildID(), tracks); err != nil {
			_, err = e.UpdateInteractionResponse(res.UpdateErr("An error occurred", err))
			return err
		}
	}

	go func() {
		if err := h.Database.AddPlayHistoryTracks(e.User().ID, tracks); err != nil {
			h.Logger.Errorf("error adding play history songs: %s", err)
		}
	}()

	_, err := e.UpdateInteractionResponse(res.UpdatePlayer(content, likeButton))
	return err
}

func (h *Handlers) OnPlayerPlayAutocomplete(e *handler.AutocompleteEvent) error {
	query := e.Data.String("query")

	limit := 24
	if strings.TrimSpace(query) == "" {
		limit = 25
	}

	tracks, err := h.Database.SearchPlay(e.User().ID, query, limit)
	if err != nil {
		h.Logger.Errorf("error searching play: %s", err)
		return e.Result(nil)
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

	return e.Result(choices)
}
