package music

import (
	"github.com/KittyBot-Org/KittyBotGo/internal/kbot"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/json"
	"github.com/disgoorg/disgolink/lavalink"
	"github.com/disgoorg/snowflake/v2"
)

var (
	_ kbot.Module         = (*module)(nil)
	_ kbot.CommandsModule = (*module)(nil)
	_ kbot.ListenerModule = (*module)(nil)
)

var Module = module{}

type module struct{}

func (m module) Commands() []kbot.Command {
	return []kbot.Command{
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
								Value: "Spotify", //string(source_extensions.SearchTypeSpotify),
							},
							{
								Name:  "Apple Music",
								Value: "Apple Music", //string(source_extensions.SearchTypeAppleMusic),
							},
						},
					},
				},
				DefaultPermission: true,
			},
			Checks: kbot.IsMemberConnectedToVoiceChannel,
			CommandHandler: map[string]kbot.CommandHandler{
				"": playHandler,
			},
			AutoCompleteHandler: map[string]kbot.AutocompleteHandler{
				"": playAutocompleteHandler,
			},
		},
		{
			Create: discord.SlashCommandCreate{
				CommandName:       "queue",
				Description:       "Shows the current queue.",
				DefaultPermission: true,
			},
			Checks: kbot.HasMusicPlayer.And(kbot.HasQueueItems),
			CommandHandler: map[string]kbot.CommandHandler{
				"": queueHandler,
			},
		},
		{
			Create: discord.SlashCommandCreate{
				CommandName:       "history",
				Description:       "Shows the current history.",
				DefaultPermission: true,
			},
			Checks: kbot.HasMusicPlayer.And(kbot.HasHistoryItems),
			CommandHandler: map[string]kbot.CommandHandler{
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
			Checks: kbot.HasMusicPlayer.And(kbot.IsMemberConnectedToVoiceChannel).And(kbot.HasQueueItems),
			CommandHandler: map[string]kbot.CommandHandler{
				"song":       removeSongHandler,
				"user-songs": removeUserSongsHandler,
			},
			AutoCompleteHandler: map[string]kbot.AutocompleteHandler{
				"song": removeSongAutocompleteHandler,
			},
		},
		{
			Create: discord.SlashCommandCreate{
				CommandName:       "clear-queue",
				Description:       "Removes all songs from your queue.",
				DefaultPermission: true,
			},
			Checks: kbot.HasMusicPlayer.And(kbot.IsMemberConnectedToVoiceChannel).And(kbot.HasQueueItems),
			CommandHandler: map[string]kbot.CommandHandler{
				"": clearQueueHandler,
			},
		},
		{
			Create: discord.SlashCommandCreate{
				CommandName:       "stop",
				Description:       "Stops the playing music.",
				DefaultPermission: true,
			},
			Checks: kbot.HasMusicPlayer.And(kbot.IsMemberConnectedToVoiceChannel),
			CommandHandler: map[string]kbot.CommandHandler{
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
								Value: int(kbot.LoopingTypeOff),
							},
							{
								Name:  "Repeat Song",
								Value: int(kbot.LoopingTypeRepeatSong),
							},
							{
								Name:  "Repeat Queue",
								Value: int(kbot.LoopingTypeRepeatQueue),
							},
						},
					},
				},
				DefaultPermission: true,
			},
			Checks: kbot.HasMusicPlayer.And(kbot.IsMemberConnectedToVoiceChannel),
			CommandHandler: map[string]kbot.CommandHandler{
				"": loopHandler,
			},
		},
		{
			Create: discord.SlashCommandCreate{
				CommandName:       "now-playing",
				Description:       "Tells you about the currently playing song.",
				DefaultPermission: true,
			},
			Checks: kbot.HasMusicPlayer.And(kbot.IsPlaying),
			CommandHandler: map[string]kbot.CommandHandler{
				"": nowPlayingHandler,
			},
			ComponentHandler: map[string]kbot.ComponentHandler{
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
			Checks: kbot.HasMusicPlayer.And(kbot.IsMemberConnectedToVoiceChannel),
			CommandHandler: map[string]kbot.CommandHandler{
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
						MinValue:    json.NewPtr(0),
						MaxValue:    json.NewPtr(100),
					},
				},
				DefaultPermission: true,
			},
			Checks: kbot.HasMusicPlayer.And(kbot.IsMemberConnectedToVoiceChannel),
			CommandHandler: map[string]kbot.CommandHandler{
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
			Checks: kbot.HasMusicPlayer.And(kbot.IsMemberConnectedToVoiceChannel),
			CommandHandler: map[string]kbot.CommandHandler{
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
						MinValue:    json.NewPtr(0),
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
			Checks: kbot.HasMusicPlayer.And(kbot.IsMemberConnectedToVoiceChannel),
			CommandHandler: map[string]kbot.CommandHandler{
				"": seekHandler,
			},
		},
		{
			Create: discord.SlashCommandCreate{
				CommandName:       "next",
				Description:       "Stops the song and starts the next one.",
				DefaultPermission: true,
			},
			Checks: kbot.HasMusicPlayer.And(kbot.IsMemberConnectedToVoiceChannel).And(kbot.HasQueueItems),
			CommandHandler: map[string]kbot.CommandHandler{
				"": nextHandler,
			},
		},
		{
			Create: discord.SlashCommandCreate{
				CommandName:       "previous",
				Description:       "Stops the song and starts the previous one.",
				DefaultPermission: true,
			},
			Checks: kbot.HasMusicPlayer.And(kbot.IsMemberConnectedToVoiceChannel).And(kbot.HasHistoryItems),
			CommandHandler: map[string]kbot.CommandHandler{
				"": previousHandler,
			},
		},
		{
			Create: discord.SlashCommandCreate{
				CommandName:       "shuffle",
				Description:       "Shuffles the queue of songs.",
				DefaultPermission: true,
			},
			Checks: kbot.HasMusicPlayer.And(kbot.IsMemberConnectedToVoiceChannel).And(kbot.HasQueueItems),
			CommandHandler: map[string]kbot.CommandHandler{
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
			CommandHandler: map[string]kbot.CommandHandler{
				"list":   likedSongsListHandler,
				"remove": likedSongsRemoveHandler,
				"clear":  likedSongsClearHandler,
				"play":   likedSongsPlayHandler,
			},
			AutoCompleteHandler: map[string]kbot.AutocompleteHandler{
				"remove": likedSongAutocompleteHandler,
				//"play":   likedSongAutocompleteHandler,
			},
		},
	}
}

func (module) OnEvent(b *kbot.Bot, event bot.Event) {
	switch e := event.(type) {
	case *events.GuildVoiceLeaveEvent:
		player := b.MusicPlayers.Get(e.VoiceState.GuildID)
		if player == nil {
			return
		}
		if e.VoiceState.UserID == b.Client.ID() {
			if err := player.Destroy(); err != nil {
				b.Logger.Error("Failed to destroy music player: ", err)
			}
			b.MusicPlayers.Delete(e.VoiceState.GuildID)
			return
		}
		if e.VoiceState.ChannelID == nil && e.OldVoiceState.ChannelID != nil {
			botVoiceState, ok := b.Client.Caches().VoiceStates().Get(e.VoiceState.GuildID, e.Client().ID())
			if ok && botVoiceState.ChannelID != nil && *botVoiceState.ChannelID == *e.OldVoiceState.ChannelID {
				voiceStates := e.Client().Caches().VoiceStates().FindAll(func(groupID snowflake.ID, voiceState discord.VoiceState) bool {
					return voiceState.ChannelID != nil && *voiceState.ChannelID == *botVoiceState.ChannelID
				})
				if len(voiceStates) == 0 {
					go player.PlanDisconnect()
				}
			}
			return
		}

	}
}
