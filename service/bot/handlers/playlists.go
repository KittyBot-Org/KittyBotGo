package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/disgolink/v2/lavalink"

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

		content += fmt.Sprintf("\nAdded %d songs to the queue from playlist `%s`", len(tracks), playlist.Name)
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

	if !urlPattern.MatchString(query) && !searchPattern.MatchString(query) {
		if source, ok := data.OptString("source"); ok {
			query = lavalink.SearchType(source).Apply(query)
		} else {
			query = lavalink.SearchTypeYouTube.Apply(query)
		}
	}

	if err := e.DeferCreateMessage(false); err != nil {
		return err
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		result, err := h.Lavalink.BestNode().Rest().LoadTracks(ctx, query)
		if err != nil {
			_, err = e.UpdateInteractionResponse(res.UpdateErr("Failed to load song", err))
			return
		}
		if result.LoadType == lavalink.LoadTypeLoadFailed {
			_, err = e.UpdateInteractionResponse(res.UpdateErr("Failed to like song", err))
		} else if result.LoadType == lavalink.LoadTypeNoMatches || len(result.Tracks) == 0 {
			_, err = e.UpdateInteractionResponse(res.UpdateError("Failed to like song: No matches found."))
		}
		if err != nil {
			h.Logger.Errorf("error loading songs: %s", err)
			return
		}
		tracks := result.Tracks
		if result.LoadType == lavalink.LoadTypeSearchResult {
			tracks = tracks[:1]
		}

		if err = h.Database.AddTracksToPlaylist(playlistID, tracks); err != nil {
			_, _ = e.UpdateInteractionResponse(res.UpdateErr("Failed to add song to playlist", err))
			return
		}
		_, _ = e.UpdateInteractionResponse(res.Updatef("Added `%d` songs to playlist", len(tracks)))
	}()

	return nil
}

func (h *Handlers) OnPlaylistRemove(e *handler.CommandEvent) error {
	trackID := e.SlashCommandInteractionData().Int("song")

	if err := h.Database.RemoveTrackFromPlaylist(trackID); err != nil {
		return e.CreateMessage(res.CreateErr("Failed to remove song from playlist", err))
	}

	return e.CreateMessage(res.Create("Removed song from playlist"))
}

func (h *Handlers) OnPlaylistRemoveAutocomplete(e *handler.AutocompleteEvent) error {
	option, ok := e.Data.Option("playlist")
	if ok && option.Focused {
		return h.OnPlaylistAutocomplete(e)
	}

	option, ok = e.Data.Option("song")
	if !ok || !option.Focused {
		return e.Result(nil)
	}

	playlistID := e.Data.Int("playlist")
	track := e.Data.String("song")

	tracks, err := h.Database.SearchPlaylistTracks(playlistID, track, 25)
	if err != nil {
		return e.Result(nil)
	}

	choices := make([]discord.AutocompleteChoice, len(tracks))
	for i, track := range tracks {
		choices[i] = discord.AutocompleteChoiceInt{
			Name:  res.Trim(track.Track.Info.Title, 100),
			Value: track.ID,
		}
	}
	return e.Result(choices)
}

func (h *Handlers) OnPlaylistAutocomplete(e *handler.AutocompleteEvent) error {
	playlists, err := h.Database.SearchPlaylists(e.User().ID, e.Data.String("track"), 25)
	if err != nil {
		return e.Result(nil)
	}

	choices := make([]discord.AutocompleteChoice, len(playlists))
	for i, playlist := range playlists {
		choices[i] = discord.AutocompleteChoiceInt{
			Name:  res.Trim(playlist.Name, 100),
			Value: playlist.ID,
		}
	}
	return e.Result(choices)
}
