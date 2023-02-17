package handlers

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/disgolink/v2/disgolink"
	"github.com/disgoorg/disgolink/v2/lavalink"
	"github.com/disgoorg/snowflake/v2"

	"github.com/KittyBot-Org/KittyBotGo/service/bot/res"
)

var (
	urlPattern    = regexp.MustCompile("^https?://[-a-zA-Z0-9+&@#/%?=~_|!:,.;]*[-a-zA-Z0-9+&@#/%=~_|]?")
	searchPattern = regexp.MustCompile(`^(.{2})(search|isrc):(.+)`)
)

func (h *Handlers) OnPlayerPlay(e *handler.CommandEvent) error {
	data := e.SlashCommandInteractionData()
	query := data.String("query")

	if source, ok := data.OptString("source"); ok {
		query = lavalink.SearchType(source).Apply(query)
	} else {
		if !urlPattern.MatchString(query) && !searchPattern.MatchString(query) {
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

	go func() {
		var loadErr error
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		player.Node().LoadTracks(ctx, query, disgolink.NewResultHandler(
			func(track lavalink.Track) {
				loadErr = h.HandleTracks(ctx, e, *voiceState.ChannelID, track)
			},
			func(playlist lavalink.Playlist) {
				loadErr = h.HandleTracks(ctx, e, *voiceState.ChannelID, playlist.Tracks...)
			},
			func(tracks []lavalink.Track) {
				loadErr = h.HandleTracks(ctx, e, *voiceState.ChannelID, tracks[0])
			},
			func() {
				_, loadErr = e.UpdateInteractionResponse(res.UpdateError("No results found for %s", query))
			},
			func(err error) {
				_, loadErr = e.UpdateInteractionResponse(res.UpdateErr("An error occurred", err))
			},
		))
		if loadErr != nil {
			h.Logger.Errorf("error loading tracks: %s", loadErr)
		}
	}()

	return nil
}

func (h *Handlers) HandleTracks(ctx context.Context, e *handler.CommandEvent, channelID snowflake.ID, tracks ...lavalink.Track) error {
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
		content += fmt.Sprintf("\nAdded %d tracks to the queue", len(tracks))
		if err := h.Database.AddQueueTracks(*e.GuildID(), tracks); err != nil {
			_, err = e.UpdateInteractionResponse(res.UpdateErr("An error occurred", err))
			return err
		}
	}

	_, err := e.UpdateInteractionResponse(res.UpdatePlayer(content, likeButton))
	return err
}
