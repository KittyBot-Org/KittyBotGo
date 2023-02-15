package handlers

import "github.com/disgoorg/disgo/discord"

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
