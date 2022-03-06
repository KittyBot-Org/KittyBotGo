package music

import (
	"github.com/DisgoOrg/disgo/core"
	"github.com/DisgoOrg/disgo/discord"
	"github.com/DisgoOrg/disgo/json"
	"github.com/DisgoOrg/disgolink/lavalink"
	"github.com/DisgoOrg/source-extensions-plugin"
	"github.com/KittyBot-Org/KittyBotGo/internal/types"
)

var (
	_ types.Module         = (*module)(nil)
	_ types.CommandsModule = (*module)(nil)
	_ types.ListenerModule = (*module)(nil)
)

var Module = module{}

type module struct{}

func (m module) Commands() []types.Command {
	return []types.Command{
		{
			Create: discord.SlashCommandCreate{
				CommandName: "play",
				Description: "Plays music for you.",
				Options: []discord.ApplicationCommandOption{
					discord.ApplicationCommandOptionString{
						Name:         "query",
						Description:  "song name or url",
						Required:     true,
						Autocomplete: true,
					},
					discord.ApplicationCommandOptionString{
						Name:        "search-provider",
						Description: "where to search for the query",
						Choices: []discord.ApplicationCommandOptionChoiceString{
							{
								Name:  "YouTube",
								Value: string(lavalink.SearchTypeYoutube),
							},
							{
								Name:  "YouTube Music",
								Value: string(lavalink.SearchTypeYoutubeMusic),
							},
							{
								Name:  "SoundCloud",
								Value: string(lavalink.SearchTypeSoundCloud),
							},
							{
								Name:  "Spotify",
								Value: string(source_extensions.SearchTypeSpotify),
							},
							{
								Name:  "Apple Music",
								Value: string(source_extensions.SearchTypeAppleMusic),
							},
						},
					},
				},
				DefaultPermission: true,
			},
			Checks: types.IsMemberConnectedToVoiceChannel,
			CommandHandler: map[string]types.CommandHandler{
				"": playHandler,
			},
			AutoCompleteHandler: map[string]types.AutocompleteHandler{
				"": playAutocompleteHandler,
			},
		},
		{
			Create: discord.SlashCommandCreate{
				CommandName:       "queue",
				Description:       "Shows the current queue.",
				DefaultPermission: true,
			},
			Checks: types.HasMusicPlayer.And(types.HasQueueItems),
			CommandHandler: map[string]types.CommandHandler{
				"": queueHandler,
			},
		},
		{
			Create: discord.SlashCommandCreate{
				CommandName:       "history",
				Description:       "Shows the current history.",
				DefaultPermission: true,
			},
			Checks: types.HasMusicPlayer.And(types.HasHistoryItems),
			CommandHandler: map[string]types.CommandHandler{
				"": historyHandler,
			},
		},
		{
			Create: discord.SlashCommandCreate{
				CommandName: "remove",
				Description: "Removes songs from the queue.",
				Options: []discord.ApplicationCommandOption{
					discord.ApplicationCommandOptionSubCommand{
						Name:        "song",
						Description: "Removes a songs from the queue.",
						Options: []discord.ApplicationCommandOption{
							discord.ApplicationCommandOptionString{
								Name:         "song",
								Description:  "the song to remove",
								Required:     true,
								Autocomplete: true,
							},
						},
					},
					discord.ApplicationCommandOptionSubCommand{
						Name:        "user-songs",
						Description: "Removes all songs from a user from the queue.",
						Options: []discord.ApplicationCommandOption{
							discord.ApplicationCommandOptionUser{
								Name:        "user",
								Description: "from which user to remove the songs",
								Required:    true,
							},
						},
					},
				},
				DefaultPermission: true,
			},
			Checks: types.HasMusicPlayer.And(types.IsMemberConnectedToVoiceChannel).And(types.HasQueueItems),
			CommandHandler: map[string]types.CommandHandler{
				"song":       removeSongHandler,
				"user-songs": removeUserSongsHandler,
			},
			AutoCompleteHandler: map[string]types.AutocompleteHandler{
				"song": removeSongAutocompleteHandler,
			},
		},
		{
			Create: discord.SlashCommandCreate{
				CommandName:       "clear-queue",
				Description:       "Removes all songs from your queue.",
				DefaultPermission: true,
			},
			Checks: types.HasMusicPlayer.And(types.IsMemberConnectedToVoiceChannel).And(types.HasQueueItems),
			CommandHandler: map[string]types.CommandHandler{
				"": clearQueueHandler,
			},
		},
		{
			Create: discord.SlashCommandCreate{
				CommandName:       "stop",
				Description:       "Stops the playing music.",
				DefaultPermission: true,
			},
			Checks: types.HasMusicPlayer.And(types.IsMemberConnectedToVoiceChannel),
			CommandHandler: map[string]types.CommandHandler{
				"": stopHandler,
			},
		},
		{
			Create: discord.SlashCommandCreate{
				CommandName: "loop",
				Description: "Loops your queue.",
				Options: []discord.ApplicationCommandOption{
					discord.ApplicationCommandOptionInt{
						Name:        "looping-type",
						Description: "how to loop your queue",
						Required:    true,
						Choices: []discord.ApplicationCommandOptionChoiceInt{
							{
								Name:  "Off",
								Value: int(types.LoopingTypeOff),
							},
							{
								Name:  "Repeat Song",
								Value: int(types.LoopingTypeRepeatSong),
							},
							{
								Name:  "Repeat Queue",
								Value: int(types.LoopingTypeRepeatQueue),
							},
						},
					},
				},
				DefaultPermission: true,
			},
			Checks: types.HasMusicPlayer.And(types.IsMemberConnectedToVoiceChannel),
			CommandHandler: map[string]types.CommandHandler{
				"": loopHandler,
			},
		},
		{
			Create: discord.SlashCommandCreate{
				CommandName:       "now-playing",
				Description:       "Tells you about the currently playing song.",
				DefaultPermission: true,
			},
			Checks: types.HasMusicPlayer.And(types.IsPlaying),
			CommandHandler: map[string]types.CommandHandler{
				"": nowPlayingHandler,
			},
			ComponentHandler: map[string]types.ComponentHandler{
				"previous":   previousComponentHandler,
				"play-pause": playPauseComponentHandler,
				"next":       nextComponentHandler,
				"like":       likeComponentHandler,
			},
		},
		{
			Create: discord.SlashCommandCreate{
				CommandName:       "pause",
				Description:       "Pauses or resumes the music.",
				DefaultPermission: true,
			},
			Checks: types.HasMusicPlayer.And(types.IsMemberConnectedToVoiceChannel),
			CommandHandler: map[string]types.CommandHandler{
				"": pauseHandler,
			},
		},
		{
			Create: discord.SlashCommandCreate{
				CommandName: "volume",
				Description: "Changes the volume of the music player.",
				Options: []discord.ApplicationCommandOption{
					discord.ApplicationCommandOptionInt{
						Name:        "volume",
						Description: "the desired volume",
						Required:    true,
						MinValue:    json.NewInt(0),
						MaxValue:    json.NewInt(100),
					},
				},
				DefaultPermission: true,
			},
			Checks: types.HasMusicPlayer.And(types.IsMemberConnectedToVoiceChannel),
			CommandHandler: map[string]types.CommandHandler{
				"": volumeHandler,
			},
		},
		{
			Create: discord.SlashCommandCreate{
				CommandName: "bass-boost",
				Description: "Enables or disables bass boost of the music player.",
				Options: []discord.ApplicationCommandOption{
					discord.ApplicationCommandOptionBool{
						Name:        "enable",
						Description: "if the bass boost should be enabled or disabled",
						Required:    true,
					},
				},
				DefaultPermission: true,
			},
			Checks: types.HasMusicPlayer.And(types.IsMemberConnectedToVoiceChannel),
			CommandHandler: map[string]types.CommandHandler{
				"": bassBoostHandler,
			},
		},
		{
			Create: discord.SlashCommandCreate{
				CommandName: "seek",
				Description: "Seeks the music to a point in the queue.",
				Options: []discord.ApplicationCommandOption{
					discord.ApplicationCommandOptionInt{
						Name:        "position",
						Description: "the position to seek to in seconds(default)/minutes/hours",
						Required:    true,
						MinValue:    json.NewInt(0),
					},
					discord.ApplicationCommandOptionInt{
						Name:        "time-unit",
						Description: "in which time unit to seek",
						Required:    false,
						Choices: []discord.ApplicationCommandOptionChoiceInt{
							{
								Name:  "Seconds",
								Value: int(lavalink.Second),
							},
							{
								Name:  "Minutes",
								Value: int(lavalink.Minute),
							},
							{
								Name:  "Hours",
								Value: int(lavalink.Hour),
							},
						},
					},
				},
				DefaultPermission: true,
			},
			Checks: types.HasMusicPlayer.And(types.IsMemberConnectedToVoiceChannel),
			CommandHandler: map[string]types.CommandHandler{
				"": seekHandler,
			},
		},
		{
			Create: discord.SlashCommandCreate{
				CommandName:       "next",
				Description:       "Stops the song and starts the next one.",
				DefaultPermission: true,
			},
			Checks: types.HasMusicPlayer.And(types.IsMemberConnectedToVoiceChannel).And(types.HasQueueItems),
			CommandHandler: map[string]types.CommandHandler{
				"": nextHandler,
			},
		},
		{
			Create: discord.SlashCommandCreate{
				CommandName:       "previous",
				Description:       "Stops the song and starts the previous one.",
				DefaultPermission: true,
			},
			Checks: types.HasMusicPlayer.And(types.IsMemberConnectedToVoiceChannel).And(types.HasHistoryItems),
			CommandHandler: map[string]types.CommandHandler{
				"": previousHandler,
			},
		},
		{
			Create: discord.SlashCommandCreate{
				CommandName:       "shuffle",
				Description:       "Shuffles the queue of songs.",
				DefaultPermission: true,
			},
			Checks: types.HasMusicPlayer.And(types.IsMemberConnectedToVoiceChannel).And(types.HasQueueItems),
			CommandHandler: map[string]types.CommandHandler{
				"": shuffleHandler,
			},
		},
		{
			Create: discord.SlashCommandCreate{
				CommandName:       "liked-songs",
				Description:       "Lists/Removes/Plays a liked song.",
				DefaultPermission: true,
				Options: []discord.ApplicationCommandOption{
					discord.ApplicationCommandOptionSubCommand{
						Name:        "list",
						Description: "Lists all your liked songs.",
					},
					discord.ApplicationCommandOptionSubCommand{
						Name:        "remove",
						Description: "Removes a liked song.",
						Options: []discord.ApplicationCommandOption{
							discord.ApplicationCommandOptionString{
								Name:         "song",
								Description:  "The song to remove",
								Required:     true,
								Autocomplete: true,
							},
						},
					},
					discord.ApplicationCommandOptionSubCommand{
						Name:        "clear",
						Description: "Clears all your liked song.",
					},
					/*discord.ApplicationCommandOptionSubCommand{
						Name:        "play",
						Description: "Plays a liked song.",
						Options: []discord.ApplicationCommandOption{
							discord.ApplicationCommandOptionString{
								Name:         "song",
								Description:  "The song to play",
								Required:     false,
								Autocomplete: true,
							},
						},
					},*/
				},
			},
			CommandHandler: map[string]types.CommandHandler{
				"list":   likedSongsListHandler,
				"remove": likedSongsRemoveHandler,
				"clear":  likedSongsClearHandler,
				"play":   likedSongsPlayHandler,
			},
			AutoCompleteHandler: map[string]types.AutocompleteHandler{
				"remove": likedSongAutocompleteHandler,
				//"play":   likedSongAutocompleteHandler,
			},
		},
	}
}

func (module) OnEvent(b *types.Bot, event core.Event) {

}
