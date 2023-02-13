package commands

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/disgoorg/disgo/discord"
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

var play = discord.SlashCommandCreate{
	Name:        "play",
	Description: "Play a song",
	Options: []discord.ApplicationCommandOption{
		discord.ApplicationCommandOptionString{
			Name:        "query",
			Description: "The song or search to play",
			Required:    true,
		},
		discord.ApplicationCommandOptionString{
			Name:        "source",
			Description: "The source to search on",
			Choices: []discord.ApplicationCommandOptionChoiceString{
				{
					Name:  "YouTube",
					Value: string(lavalink.SearchTypeYouTube),
				},
				{
					Name:  "YouTube Music",
					Value: string(lavalink.SearchTypeYouTubeMusic),
				},
				{
					Name:  "SoundCloud",
					Value: string(lavalink.SearchTypeSoundCloud),
				},
				{
					Name:  "Deezer",
					Value: "dzsearch",
				},
				{
					Name:  "Deezer ISRC",
					Value: "dzisrc",
				},
			},
		},
	},
}

func (h *Cmds) OnPlay(e *handler.CommandEvent) error {
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

	if err := e.DeferCreateMessage(false); err != nil {
		return err
	}

	player := h.Lavalink.Player(*e.GuildID())

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

func (h *Cmds) HandleTracks(ctx context.Context, e *handler.CommandEvent, channelID snowflake.ID, tracks ...lavalink.Track) error {
	_, ok := h.Discord.Caches().VoiceState(*e.GuildID(), e.ApplicationID())
	if !ok {
		if err := h.Discord.UpdateVoiceState(context.Background(), *e.GuildID(), &channelID, false, false); err != nil {
			_, err = e.UpdateInteractionResponse(res.UpdateErr("An error occurred", err))
			return err
		}
	}
	player := h.Player(*e.GuildID())
	var content string
	if player.Track() == nil {
		track := tracks[0]
		tracks = tracks[1:]

		if err := player.Update(ctx, lavalink.WithTrack(track)); err != nil {
			_, err = e.UpdateInteractionResponse(res.UpdateErr("An error occurred", err))
			return err
		}
		content = fmt.Sprintf("Playing %s", res.FormatTrack(track, 0))
	}

	if len(tracks) > 0 {
		content += fmt.Sprintf("Added %d tracks to the queue", len(tracks))
		player.Queue.Add(tracks)
	}

	_, err := e.UpdateInteractionResponse(res.Update(content))
	return err
}
