package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/disgolink/v2/disgolink"
	"github.com/disgoorg/disgolink/v2/lavalink"
	"github.com/lithammer/fuzzysearch/fuzzy"

	"github.com/KittyBot-Org/KittyBotGo/service/bot/res"
)

var playlistsCommand = discord.SlashCommandCreate{
	Name:        "playlists",
	Description: "Lets you manage your playlists",
	Options: []discord.ApplicationCommandOption{
		discord.ApplicationCommandOptionSubCommand{
			Name:        "list",
			Description: "List all your playlists",
		},
		discord.ApplicationCommandOptionSubCommand{
			Name:        "show",
			Description: "Shows a playlist",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionInt{
					Name:         "playlist",
					Description:  "The name of the playlist",
					Required:     true,
					Autocomplete: true,
				},
			},
		},
		discord.ApplicationCommandOptionSubCommand{
			Name:        "create",
			Description: "Create a new playlist",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionString{
					Name:        "name",
					Description: "The name of the playlist",
					Required:    true,
				},
			},
		},
		discord.ApplicationCommandOptionSubCommand{
			Name:        "delete",
			Description: "Delete a playlist",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionInt{
					Name:         "playlist",
					Description:  "The name of the playlist",
					Required:     true,
					Autocomplete: true,
				},
			},
		},
		discord.ApplicationCommandOptionSubCommand{
			Name:        "add",
			Description: "Add a song to a playlist",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionInt{
					Name:         "playlist",
					Description:  "The name of the playlist",
					Required:     true,
					Autocomplete: true,
				},
				discord.ApplicationCommandOptionString{
					Name:        "query",
					Description: "The link or query of the song",
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
		},
		discord.ApplicationCommandOptionSubCommand{
			Name:        "play",
			Description: "Plays a playlist",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionInt{
					Name:         "playlist",
					Description:  "The name of the playlist",
					Required:     true,
					Autocomplete: true,
				},
			},
		},
		discord.ApplicationCommandOptionSubCommand{
			Name:        "remove",
			Description: "Remove a song from a playlist",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionInt{
					Name:         "playlist",
					Description:  "The name of the playlist",
					Required:     true,
					Autocomplete: true,
				},
				discord.ApplicationCommandOptionInt{
					Name:         "song",
					Description:  "The name of the song",
					Required:     true,
					Autocomplete: true,
				},
			},
		},
	},
}

func (h *Handlers) OnPlaylistsList(e *handler.CommandEvent) error {
	playlists, err := h.Database.GetPlaylists(e.User().ID)
	if err != nil {
		return e.CreateMessage(res.CreateErr("Failed to get playlists", err))
	}

	if len(playlists) == 0 {
		return e.CreateMessage(res.CreateError("You have no playlists"))
	}

	content := fmt.Sprintf("Your playlists(%d):\n", len(playlists))
	for i, playlist := range playlists {
		content += fmt.Sprintf("%d: %s\n", i+1, playlist.Name)
	}

	return e.CreateMessage(res.Create(content))
}

func (h *Handlers) OnPlaylistCreate(e *handler.CommandEvent) error {
	data := e.SlashCommandInteractionData()
	playlist, err := h.Database.CreatePlaylist(e.User().ID, data.String("name"))
	if err != nil {
		return e.CreateMessage(res.CreateErr("Failed to create playlist", err))
	}

	return e.CreateMessage(res.Create(fmt.Sprintf("Created playlist: `%s`", playlist.Name)))
}

func (h *Handlers) OnPlaylistDelete(e *handler.CommandEvent) error {
	data := e.SlashCommandInteractionData()
	if err := h.Database.DeletePlaylist(data.Int("playlist"), e.User().ID); err != nil {
		return e.CreateMessage(res.CreateErr("Failed to delete playlist", err))
	}

	return e.CreateMessage(res.Create(fmt.Sprintf("Deleted playlist")))
}

func (h *Handlers) OnPlaylistShow(e *handler.CommandEvent) error {
	data := e.SlashCommandInteractionData()
	playlist, tracks, err := h.Database.GetPlaylist(data.Int("playlist"))
	if err != nil {
		return e.CreateMessage(res.CreateErr("Failed to get playlist", err))
	}

	if len(tracks) == 0 {
		return e.CreateMessage(res.CreateError("Playlist is empty"))
	}

	content := fmt.Sprintf("Playlist `%s`(%d):\n", playlist.Name, len(tracks))
	for i, track := range tracks {
		line := fmt.Sprintf("%d. %s\n", i+1, res.FormatTrack(track.Track, 0))
		if len([]rune(content))+len([]rune(line)) > 2000 {
			break
		}
		content += line
	}

	return e.CreateMessage(res.Create(content))
}

func (h *Handlers) OnPlaylistPlay(e *handler.CommandEvent) error {
	data := e.SlashCommandInteractionData()

	voiceState, ok := h.Discord.Caches().VoiceState(*e.GuildID(), e.User().ID)
	if !ok {
		return e.CreateMessage(res.CreateError("You are not in a voice channel"))
	}

	playlist, dbTracks, err := h.Database.GetPlaylist(data.Int("playlist"))
	if err != nil {
		return e.CreateMessage(res.CreateErr("Failed to get playlist", err))
	}

	if len(dbTracks) == 0 {
		return e.CreateMessage(res.CreateError("Playlist is empty"))
	}

	_, ok = h.Discord.Caches().VoiceState(*e.GuildID(), e.ApplicationID())
	if !ok {
		if err = h.Discord.UpdateVoiceState(context.Background(), *e.GuildID(), voiceState.ChannelID, false, false); err != nil {
			_, err = e.UpdateInteractionResponse(res.UpdateErr("Failed to join channel", err))
			return err
		}
	}

	player := h.Lavalink.Player(*e.GuildID())
	if _, err = h.Database.GetPlayer(*e.GuildID(), player.Node().Config().Name); err != nil {
		return e.CreateMessage(res.CreateErr("Failed to get or create player", err))
	}

	var content string
	if player.Track() == nil {
		track := dbTracks[0]
		dbTracks = dbTracks[1:]

		if err = player.Update(context.Background(), lavalink.WithTrack(track.Track)); err != nil {
			return e.CreateMessage(res.CreateErr("An error occurred", err))
		}
		content = fmt.Sprintf("▶ Playing: %s from playlist `%s`", res.FormatTrack(track.Track, 0), playlist.Name)
	}

	if len(dbTracks) > 0 {
		tracks := make([]lavalink.Track, len(dbTracks))
		for i := range dbTracks {
			tracks[i] = dbTracks[i].Track
		}

		content += fmt.Sprintf("\nAdded %d tracks to the queue from playlist `%s`", len(tracks), playlist.Name)
		if err = h.Database.AddQueueTracks(*e.GuildID(), tracks); err != nil {
			return e.CreateMessage(res.CreateErr("An error occurred", err))
		}
	}

	return e.CreateMessage(res.Create(content))
}

func (h *Handlers) OnPlaylistAdd(e *handler.CommandEvent) error {
	data := e.SlashCommandInteractionData()
	playlistID := data.Int("playlist")
	query := data.String("query")

	if source, ok := data.OptString("source"); ok {
		query = lavalink.SearchType(source).Apply(query)
	} else {
		if !urlPattern.MatchString(query) && !searchPattern.MatchString(query) {
			query = lavalink.SearchTypeYouTube.Apply(query)
		}
	}

	if err := e.DeferCreateMessage(false); err != nil {
		return err
	}

	go func() {
		var loadErr error
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		h.Lavalink.BestNode().LoadTracks(ctx, query, disgolink.NewResultHandler(
			func(track lavalink.Track) {
				loadErr = h.handlePlaylistTracks(e, playlistID, track)
			},
			func(playlist lavalink.Playlist) {
				loadErr = h.handlePlaylistTracks(e, playlistID, playlist.Tracks...)
			},
			func(tracks []lavalink.Track) {
				loadErr = h.handlePlaylistTracks(e, playlistID, tracks[0])
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

func (h *Handlers) handlePlaylistTracks(e *handler.CommandEvent, playlistID int, tracks ...lavalink.Track) error {
	if err := h.Database.AddTracksToPlaylist(playlistID, tracks); err != nil {
		_, err = e.UpdateInteractionResponse(res.UpdateErr("Failed to add track to playlist", err))
		return err
	}

	if len(tracks) == 1 {
		_, err := e.UpdateInteractionResponse(res.Updatef("Added track to playlist: %s", res.FormatTrack(tracks[0], 0)))
		return err
	}

	_, err := e.UpdateInteractionResponse(res.Updatef("Added `%d` tracks to playlist", len(tracks)))
	return err
}

func (h *Handlers) OnPlaylistAutocomplete(e *handler.AutocompleteEvent) error {
	playlists, err := h.Database.GetPlaylists(e.User().ID)
	if err != nil {
		return e.Result(nil)
	}

	playlistValues := make(map[string]int, len(playlists))
	playlistNames := make([]string, len(playlists))
	for i, playlist := range playlists {
		name := trim(playlist.Name, 100)
		playlistValues[name] = playlist.ID
		playlistNames[i] = name
	}

	ranks := fuzzy.RankFindFold(e.Data.String("track"), playlistNames)
	choicesLen := len(ranks)
	if choicesLen > 25 {
		choicesLen = 25
	}
	choices := make([]discord.AutocompleteChoice, choicesLen)
	for i, rank := range ranks {
		if i >= 25 {
			break
		}
		choices[i] = discord.AutocompleteChoiceInt{
			Name:  rank.Target,
			Value: playlistValues[rank.Target],
		}
	}
	return e.Result(choices)
}
