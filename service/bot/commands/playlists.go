package commands

import (
	"context"
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/disgolink/v3/lavalink"
	"github.com/disgoorg/lavaqueue-plugin"

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

func (c *commands) OnPlaylistsList(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	playlists, err := c.Database.GetPlaylists(e.User().ID)
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

func (c *commands) OnPlaylistCreate(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	playlist, err := c.Database.CreatePlaylist(e.User().ID, data.String("name"))
	if err != nil {
		return e.CreateMessage(res.CreateErr("Failed to create playlist", err))
	}

	return e.CreateMessage(res.Create(fmt.Sprintf("Created playlist: `%s`", playlist.Name)))
}

func (c *commands) OnPlaylistDelete(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	if err := c.Database.DeletePlaylist(data.Int("playlist"), e.User().ID); err != nil {
		return e.CreateMessage(res.CreateErr("Failed to delete playlist", err))
	}

	return e.CreateMessage(res.Create(fmt.Sprintf("Deleted playlist")))
}

func (c *commands) OnPlaylistShow(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	playlist, tracks, err := c.Database.GetPlaylist(data.Int("playlist"))
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

func (c *commands) OnPlaylistPlay(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	voiceState, ok := c.Discord.Caches().VoiceState(*e.GuildID(), e.User().ID)
	if !ok {
		return e.CreateMessage(res.CreateError("You are not in a voice channel"))
	}

	playlist, dbTracks, err := c.Database.GetPlaylist(data.Int("playlist"))
	if err != nil {
		return e.CreateMessage(res.CreateErr("Failed to get playlist", err))
	}

	if len(dbTracks) == 0 {
		return e.CreateMessage(res.CreateError("Playlist is empty"))
	}

	_, ok = c.Discord.Caches().VoiceState(*e.GuildID(), e.ApplicationID())
	if !ok {
		if err = c.Discord.UpdateVoiceState(context.Background(), *e.GuildID(), voiceState.ChannelID, false, false); err != nil {
			return e.CreateMessage(res.CreateErr("Failed to join channel", err))
		}
	}

	queueTracks := make([]lavaqueue.QueueTrack, len(dbTracks))
	for i, track := range dbTracks {
		queueTracks[i] = lavaqueue.QueueTrack{
			Encoded:  track.Track.Encoded,
			UserData: nil, // TODO: Add user data
		}
	}

	player := c.Lavalink.Player(*e.GuildID())
	track, err := lavaqueue.AddQueueTracks(e.Ctx, player.Node(), *e.GuildID(), queueTracks)
	if err != nil {
		return e.CreateMessage(res.CreateErr("An error occurred playing the song", err))
	}

	var (
		content     string
		likeButton  bool
		tracksCount = len(dbTracks)
	)
	if track != nil {
		content = fmt.Sprintf("▶ Playing: %s from playlist `%s`", res.FormatTrack(*track, 0), playlist.Name)
		likeButton = true
		tracksCount--
	}
	if len(dbTracks) > 0 {
		content += fmt.Sprintf("\nAdded %d songs to the queue from playlist `%s`", tracksCount, playlist.Name)
	}

	return e.CreateMessage(res.CreatePlayer(content, likeButton))
}

func (c *commands) OnPlaylistAdd(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
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

	result, err := c.Lavalink.BestNode().Rest().LoadTracks(context.Background(), query)
	if err != nil {
		_, err = e.UpdateInteractionResponse(res.UpdateErr("Failed to load song", err))
		return err
	}

	var tracks []lavalink.Track
	switch d := result.Data.(type) {
	case lavalink.Exception:
		_, err = e.UpdateInteractionResponse(res.UpdateErr("Failed to find song", err))
		return err
	case lavalink.Empty:
		_, err = e.UpdateInteractionResponse(res.UpdateError("Failed to find song: No matches found."))
		return err
	case lavalink.Track:
		tracks = append(tracks, d)
	case lavalink.Search:
		if len(d) == 0 {
			_, err = e.UpdateInteractionResponse(res.UpdateError("Failed to find song: No matches found."))
			return err
		}
		tracks = d[:1]
	case lavalink.Playlist:
		tracks = d.Tracks
	}

	if err = c.Database.AddTracksToPlaylist(playlistID, tracks); err != nil {
		_, err = e.UpdateInteractionResponse(res.UpdateErr("Failed to add song to playlist", err))
		return err
	}
	_, err = e.UpdateInteractionResponse(res.Updatef("Added `%d` songs to playlist", len(tracks)))
	return err
}

func (c *commands) OnPlaylistRemove(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	trackID := data.Int("song")

	if err := c.Database.RemoveTrackFromPlaylist(trackID); err != nil {
		return e.CreateMessage(res.CreateErr("Failed to remove song from playlist", err))
	}

	return e.CreateMessage(res.Create("Removed song from playlist"))
}

func (c *commands) OnPlaylistRemoveAutocomplete(e *handler.AutocompleteEvent) error {
	option, ok := e.Data.Option("playlist")
	if ok && option.Focused {
		return c.OnPlaylistAutocomplete(e)
	}

	option, ok = e.Data.Option("song")
	if !ok || !option.Focused {
		return e.AutocompleteResult(nil)
	}

	playlistID := e.Data.Int("playlist")
	track := e.Data.String("song")

	tracks, err := c.Database.SearchPlaylistTracks(playlistID, track, 25)
	if err != nil {
		return e.AutocompleteResult(nil)
	}

	choices := make([]discord.AutocompleteChoice, len(tracks))
	for i, track := range tracks {
		choices[i] = discord.AutocompleteChoiceInt{
			Name:  res.Trim(track.Track.Info.Title, 100),
			Value: track.ID,
		}
	}
	return e.AutocompleteResult(choices)
}

func (c *commands) OnPlaylistAutocomplete(e *handler.AutocompleteEvent) error {
	playlists, err := c.Database.SearchPlaylists(e.User().ID, e.Data.String("track"), 25)
	if err != nil {
		return e.AutocompleteResult(nil)
	}

	choices := make([]discord.AutocompleteChoice, len(playlists))
	for i, playlist := range playlists {
		choices[i] = discord.AutocompleteChoiceInt{
			Name:  res.Trim(playlist.Name, 100),
			Value: playlist.ID,
		}
	}
	return e.AutocompleteResult(choices)
}
