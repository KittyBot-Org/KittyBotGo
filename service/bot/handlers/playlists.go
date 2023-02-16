package handlers

import (
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
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
					Name:         "song",
					Description:  "The name of the song",
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
