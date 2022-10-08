package commands

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/KittyBot-Org/KittyBotGo/dbot"
	"github.com/KittyBot-Org/KittyBotGo/dbot/responses"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgolink/lavalink"
	"github.com/disgoorg/handler"
	"github.com/disgoorg/snowflake/v2"
	source_plugins "github.com/disgoorg/source-plugins"
	"github.com/lithammer/fuzzysearch/fuzzy"
)

var urlPattern = regexp.MustCompile("^https?://[-a-zA-Z0-9+&@#/%?=~_|!:,.;]*[-a-zA-Z0-9+&@#/%=~_|]?")
var trackRegex = regexp.MustCompile(`\[\x60(?P<title>.+)\x60]\((?P<url>.+)?\)`)
var searchPattern = regexp.MustCompile(`^(.{2})search:(.+)`)

func Play(b *dbot.Bot) handler.Command {
	return handler.Command{
		Create: discord.SlashCommandCreate{
			Name:        "play",
			Description: "Plays music for you.",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionString{
					Name:         "query",
					Description:  "Song name or url",
					Required:     true,
					Autocomplete: true,
				},
				discord.ApplicationCommandOptionString{
					Name:        "search-provider",
					Description: "Where to search for the query",
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
							Value: string(source_plugins.SearchTypeSpotify),
						},
						{
							Name:  "Apple Music",
							Value: string(source_plugins.SearchTypeAppleMusic),
						},
						{
							Name:  "Deezer ISRC",
							Value: string(source_plugins.SearchTypeDeezerISRC),
						},
						{
							Name:  "Deezer",
							Value: string(source_plugins.SearchTypeDeezer),
						},
					},
				},
			},
		},
		Check: dbot.IsMemberConnectedToVoiceChannel(b),
		CommandHandlers: map[string]handler.CommandHandler{
			"": playHandler(b),
		},
		AutocompleteHandlers: map[string]handler.AutocompleteHandler{
			"": playAutocompleteHandler(b),
		},
	}
}

func playHandler(b *dbot.Bot) handler.CommandHandler {
	return func(e *events.ApplicationCommandInteractionCreate) error {
		data := e.SlashCommandInteractionData()

		if voiceState, ok := b.Client.Caches().VoiceStates().Get(*e.GuildID(), e.User().ID); !ok || voiceState.ChannelID == nil {
			return e.CreateMessage(responses.CreateErrorf("You must be in a voice channel to use this command."))
		}

		channel, _ := e.GuildChannel()
		selfMember, _ := e.Client().Caches().GetSelfMember(*e.GuildID())
		if perms := b.Client.Caches().GetMemberPermissionsInChannel(channel, selfMember); perms.Missing(discord.PermissionVoiceConnect) {
			return e.CreateMessage(responses.CreateErrorf("It seems like I don't have permissions to join your voice channel."))
		}

		query := data.String("query")
		if searchProvider, ok := data.OptString("search-provider"); ok {
			query = lavalink.SearchType(searchProvider).Apply(query)
		} else {
			if !urlPattern.MatchString(query) && !searchPattern.MatchString(query) {
				query = lavalink.SearchTypeYoutube.Apply(query)
			}
		}

		if err := e.DeferCreateMessage(false); err != nil {
			return err
		}
		err := b.Lavalink.BestRestClient().LoadItemHandler(context.TODO(), query, lavalink.NewResultHandler(
			func(track lavalink.AudioTrack) {
				if err := b.DB.PlayHistory().Add(e.User().ID, query, track.Info().Title); err != nil {
					b.Logger.Error("Failed to add track to play history: ", err)
				}
				playAndQueue(b, e.BaseInteraction, track)
			},
			func(playlist lavalink.AudioPlaylist) {
				if err := b.DB.PlayHistory().Add(e.User().ID, query, playlist.Name()); err != nil {
					b.Logger.Error("Failed to add track to play history: ", err)
				}
				playAndQueue(b, e.BaseInteraction, playlist.Tracks()...)
			},
			func(tracks []lavalink.AudioTrack) {
				if err := b.DB.PlayHistory().Add(e.User().ID, query, data.String("query")); err != nil {
					b.Logger.Error("Failed to add track to play history: ", err)
				}
				giveSearchSelection(b, e, tracks)
			},
			func() {
				if _, err := e.Client().Rest().UpdateInteractionResponse(e.ApplicationID(), e.Token(), responses.UpdateErrorf("No track/s found for your link/query")); err != nil {
					b.Logger.Error("Failed to update not found message: ", err)
				}
			},
			func(ex lavalink.FriendlyException) {
				if _, err := e.Client().Rest().UpdateInteractionResponse(e.ApplicationID(), e.Token(), responses.UpdateErrorf("Failed to load track/s: %s", ex.Message)); err != nil {
					b.Logger.Error("Failed to update error message: ", err)
				}
			},
		))
		if err != nil {
			if _, err = e.Client().Rest().UpdateInteractionResponse(e.ApplicationID(), e.Token(), responses.UpdateErrorf("Failed to lookup track. Please try again.")); err != nil {
				b.Logger.Error("Failed to update error message: ", err)
			}
		}
		return err
	}
}

func playAutocompleteHandler(b *dbot.Bot) handler.AutocompleteHandler {
	return func(e *events.AutocompleteInteractionCreate) error {
		query := e.Data.String("query")
		playHistory, err := b.DB.PlayHistory().Get(e.User().ID)
		if err != nil {
			b.Logger.Error("Error adding music history entry: ", err)
			return err
		}
		likedSongs, err := b.DB.LikedSongs().GetAll(e.User().ID)
		if err != nil {
			b.Logger.Error("Failed to get music history entries: ", err)
			return err
		}
		if (len(playHistory)+len(likedSongs) == 0) && query == "" {
			return e.Result(nil)
		}

		labels := make([]string, len(playHistory)+len(likedSongs))
		unsortedResult := make(map[string]string, len(playHistory)+len(likedSongs))
		i := 0
		for _, entry := range playHistory {
			title := "🔁" + entry.Title
			unsortedResult[title] = entry.Query
			labels[i] = title
			i++
		}

		for _, entry := range likedSongs {
			title := "❤" + entry.Title
			unsortedResult[title] = entry.Query
			labels[i] = title
			i++
		}

		if query == "" {
			var choices []discord.AutocompleteChoice
			for key, value := range unsortedResult {
				choices = append(choices, discord.AutocompleteChoiceString{
					Name:  key,
					Value: value,
				})
			}
			return e.Result(choices)
		}

		ranks := fuzzy.RankFindFold(query, labels)
		resultLen := len(ranks)
		if resultLen > 24 {
			resultLen = 24
		}
		result := make([]discord.AutocompleteChoice, resultLen+1)
		queryEmoji := "🔎"
		if urlPattern.MatchString(query) {
			queryEmoji = "🔗"
		}
		result[0] = discord.AutocompleteChoiceString{
			Name:  queryEmoji + query,
			Value: query,
		}
		for ii, rank := range ranks {
			if ii >= resultLen {
				break
			}
			result[ii+1] = discord.AutocompleteChoiceString{
				Name:  rank.Target,
				Value: unsortedResult[rank.Target],
			}
		}
		return e.Result(result)
	}
}

func playAndQueue(b *dbot.Bot, i discord.BaseInteraction, tracks ...lavalink.AudioTrack) {
	player := b.MusicPlayers.Get(*i.GuildID())
	if player == nil {
		player = b.MusicPlayers.New(*i.GuildID(), dbot.PlayerTypeMusic, dbot.LoopingTypeOff)
		b.MusicPlayers.Add(player)
	}
	var voiceChannelID snowflake.ID
	if voiceState, ok := b.Client.Caches().VoiceStates().Get(*i.GuildID(), i.User().ID); !ok || voiceState.ChannelID == nil {
		if _, err := b.Client.Rest().UpdateInteractionResponse(i.ApplicationID(), i.Token(), responses.UpdateErrorComponentsf("You need to be in a voice channel.", nil)); err != nil {
			b.Logger.Error("Failed to update error message: ", err)
		}
		return
	} else {
		voiceChannelID = *voiceState.ChannelID
	}

	if voiceState, ok := b.Client.Caches().VoiceStates().Get(*i.GuildID(), b.Client.ID()); !ok || voiceState.ChannelID == nil || *voiceState.ChannelID != voiceChannelID {
		if err := b.Client.Connect(context.TODO(), *i.GuildID(), voiceChannelID); err != nil {
			if _, err = b.Client.Rest().UpdateInteractionResponse(i.ApplicationID(), i.Token(), responses.UpdateErrorComponentsf("Failed to connect to your voice channel. Please try again.", nil)); err != nil {
				b.Logger.Error("Failed to update error message: ", err)
			}
			return
		}
	}

	for ii := range tracks {
		tracks[ii].SetUserData(dbot.AudioTrackData{
			Requester: i.User().ID,
		})
	}

	fmt.Printf("PlayingTrack: %v", player.PlayingTrack())

	if player.PlayingTrack() == nil {
		track := tracks[0]
		if len(tracks) > 0 {
			tracks = tracks[1:]
		}
		if err := player.Play(track); err != nil {
			if _, err = b.Client.Rest().UpdateInteractionResponse(i.ApplicationID(), i.Token(), responses.UpdateErrorComponentsf("Failed to play song. Please try again.", nil)); err != nil {
				b.Logger.Error("Error while playing song: ", err)
			}
			return
		}
		if _, err := b.Client.Rest().UpdateInteractionResponse(i.ApplicationID(), i.Token(), responses.UpdateSuccessComponentsf("▶️ Now playing: %s", []any{formatTrack(track)}, getMusicControllerComponents(track))); err != nil {
			b.Logger.Error("Error while updating interaction message: ", err)
		}
	} else {
		if _, err := b.Client.Rest().UpdateInteractionResponse(i.ApplicationID(), i.Token(), responses.UpdateErrorComponentsf("Added `%d` songs to the queue.", []any{len(tracks)}, getMusicControllerComponents(nil))); err != nil {
			b.Logger.Error("Error while updating interaction message: ", err)
		}
	}
	if len(tracks) > 0 {
		player.Queue.Push(tracks...)
	}
}

func giveSearchSelection(b *dbot.Bot, e *events.ApplicationCommandInteractionCreate, tracks []lavalink.AudioTrack) {
	var options []discord.SelectMenuOption
	for i, track := range tracks {
		if len(options) >= 25 {
			break
		}
		label := track.Info().Title
		if len(label) > 80 {
			label = label[:79] + "…"
		}
		description := "by: " + track.Info().Author
		if len(description) > 100 {
			description = description[:99] + "…"
		}

		options = append(options, discord.SelectMenuOption{
			Label:       label,
			Description: description,
			Value:       strconv.Itoa(i),
		})
	}
	if _, err := e.Client().Rest().UpdateInteractionResponse(e.ApplicationID(), e.Token(), responses.UpdateErrorComponentsf("Select songs to play.", []any{len(tracks)})); err != nil {
		b.Logger.Error("Error while updating interaction message: ", err)
	}

	if _, err := e.Client().Rest().UpdateInteractionResponse(e.ApplicationID(), e.Token(),
		responses.UpdateSuccessComponentsf("Select songs to play.", nil, discord.NewActionRow(
			discord.NewSelectMenu("search:"+e.ID().String(), "Select songs to play.", options...).WithMaxValues(len(options)),
		)),
	); err != nil {
		b.Logger.Error("Error while updating interaction message: ", err)
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		bot.WaitForEvent(e.Client(), ctx,
			func(ne *events.ComponentInteractionCreate) bool {
				if ne.Data.CustomID() == "search:"+e.ID().String() {
					if e.User().ID == ne.User().ID {
						return true
					}
					err := ne.CreateMessage(responses.CreateErrorf("You can't select songs for someone else."))
					if err != nil {
						b.Logger.Error("Error while creating message: ", err)
					}
				}
				return false
			},
			func(ne *events.ComponentInteractionCreate) {
				if err := ne.DeferUpdateMessage(); err != nil {
					b.Logger.Error(err)
					return
				}
				var playTracks []lavalink.AudioTrack
				for _, value := range ne.SelectMenuInteractionData().Values {
					index, _ := strconv.Atoi(value)
					playTracks = append(playTracks, tracks[index])
				}
				playAndQueue(b, e.BaseInteraction, playTracks...)
				return
			},
			func() {
				if _, err := e.Client().Rest().UpdateInteractionResponse(e.ApplicationID(), e.Token(), responses.UpdateErrorComponentsf("Search timed out after 60s. Please try again.", nil)); err != nil {
					b.Logger.Error("Error while updating interaction message: ", err)
				}
			},
		)
	}()
}
