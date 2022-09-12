package commands

import (
	"context"
	"regexp"
	"strconv"
	"time"

	"github.com/KittyBot-Org/KittyBotGo/dbot"
	"github.com/KittyBot-Org/KittyBotGo/dbot/responses"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgolink/lavalink"
	"github.com/disgoorg/snowflake/v2"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"golang.org/x/text/message"
)

var urlPattern = regexp.MustCompile("^https?://[-a-zA-Z0-9+&@#/%?=~_|!:,.;]*[-a-zA-Z0-9+&@#/%=~_|]?")
var trackRegex = regexp.MustCompile(`\[\x60(?P<title>.+)\x60]\((?P<url>.+)?\)`)
var searchPattern = regexp.MustCompile(`^(.{2})search:(.+)`)

var Play = handler.Command{
	Create: discord.SlashCommandCreate{
		Name:        "play",
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
	},
	Check: dbot.IsMemberConnectedToVoiceChannel,
	CommandHandlers: map[string]handler.CommandHandler{
		"": playHandler,
	},
	AutoCompleteHandler: map[string]dbot.AutocompleteHandler{
		"": playAutocompleteHandler,
	},
}

func playHandler(b *dbot.Bot, p *message.Printer, e *events.ApplicationCommandInteractionCreate) error {
	data := e.SlashCommandInteractionData()

	var voiceChannelID snowflake.ID
	if voiceState, ok := b.Client.Caches().VoiceStates().Get(*e.GuildID(), e.User().ID); !ok || voiceState.ChannelID == nil {
		return e.CreateMessage(responses.CreateErrorf(p, "modules.music.not.in.voice"))
	} else {
		voiceChannelID = *voiceState.ChannelID
	}

	channel, _ := e.GuildChannel()
	if perms := b.Client.Caches().GetMemberPermissionsInChannel(channel, e.Member().Member); perms.Missing(discord.PermissionVoiceConnect) {
		return e.CreateMessage(responses.CreateErrorf(p, "modules.music.missing.perms"))
	}

	if voiceState, ok := b.Client.Caches().VoiceStates().Get(*e.GuildID(), b.Client.ID()); !ok || voiceState.ChannelID == nil || *voiceState.ChannelID != voiceChannelID {
		if err := b.Client.Connect(context.TODO(), *e.GuildID(), voiceChannelID); err != nil {
			return e.CreateMessage(responses.CreateErrorf(p, "modules.music.connect.error"))
		}
	}

	query := data.String("query")
	if searchProvider, ok := data.OptString("searchProvider"); ok {
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
			playAndQueue(b, p, e.BaseInteraction, track)
		},
		func(playlist lavalink.AudioPlaylist) {
			if err := b.DB.PlayHistory().Add(e.User().ID, query, playlist.Name()); err != nil {
				b.Logger.Error("Failed to add track to play history: ", err)
			}
			playAndQueue(b, p, e.BaseInteraction, playlist.Tracks()...)
		},
		func(tracks []lavalink.AudioTrack) {
			if err := b.DB.PlayHistory().Add(e.User().ID, query, data.String("query")); err != nil {
				b.Logger.Error("Failed to add track to play history: ", err)
			}
			giveSearchSelection(b, p, e, tracks)
		},
		func() {
			if _, err := e.Client().Rest().UpdateInteractionResponse(e.ApplicationID(), e.Token(), responses.UpdateErrorf(p, "modules.music.commands.play.no.results")); err != nil {
				b.Logger.Error("Failed to update not found message: ", err)
			}
		},
		func(ex lavalink.FriendlyException) {
			if _, err := e.Client().Rest().UpdateInteractionResponse(e.ApplicationID(), e.Token(), responses.UpdateErrorf(p, "modules.music.commands.play.error", ex.Message)); err != nil {
				b.Logger.Error("Failed to update error message: ", err)
			}
		},
	))
	if err != nil {
		if _, err = e.Client().Rest().UpdateInteractionResponse(e.ApplicationID(), e.Token(), responses.UpdateErrorf(p, "modules.music.commands.play.error")); err != nil {
			b.Logger.Error("Failed to update error message: ", err)
		}
	}
	return err
}

func playAutocompleteHandler(b *dbot.Bot, _ *message.Printer, e *events.AutocompleteInteractionCreate) error {
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

func playAndQueue(b *dbot.Bot, p *message.Printer, i discord.BaseInteraction, tracks ...lavalink.AudioTrack) {
	player := b.MusicPlayers.Get(*i.GuildID())
	if player == nil {
		player = b.MusicPlayers.New(*i.GuildID(), dbot.PlayerTypeMusic, dbot.LoopingTypeOff)
		b.MusicPlayers.Add(player)
	}
	var voiceChannelID snowflake.ID
	if voiceState, ok := b.Client.Caches().VoiceStates().Get(*i.GuildID(), i.User().ID); !ok || voiceState.ChannelID == nil {
		if _, err := b.Client.Rest().UpdateInteractionResponse(i.ApplicationID(), i.Token(), responses.UpdateErrorComponentsf(p, "modules.music.not.in.voice", nil)); err != nil {
			b.Logger.Error("Failed to update error message: ", err)
		}
		return
	} else {
		voiceChannelID = *voiceState.ChannelID
	}

	if voiceState, ok := b.Client.Caches().VoiceStates().Get(*i.GuildID(), b.Client.ID()); !ok || voiceState.ChannelID == nil || *voiceState.ChannelID != voiceChannelID {
		if err := b.Client.Connect(context.TODO(), *i.GuildID(), voiceChannelID); err != nil {
			if _, err = b.Client.Rest().UpdateInteractionResponse(i.ApplicationID(), i.Token(), responses.UpdateErrorComponentsf(p, "modules.music.connect.error", nil)); err != nil {
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

	if player.PlayingTrack() == nil {
		track := tracks[0]
		if len(tracks) > 0 {
			tracks = tracks[1:]
		}
		if err := player.Play(track); err != nil {
			if _, err = b.Client.Rest().UpdateInteractionResponse(i.ApplicationID(), i.Token(), responses.UpdateErrorComponentsf(p, "modules.music.commands.play.error", nil)); err != nil {
				b.Logger.Error("Error while playing song: ", err)
			}
			return
		}
		if _, err := b.Client.Rest().UpdateInteractionResponse(i.ApplicationID(), i.Token(), responses.UpdateSuccessComponentsf(p, "modules.music.commands.play.now.playing", []any{formatTrack(track)}, getMusicControllerComponents(track))); err != nil {
			b.Logger.Error("Error while updating interaction message: ", err)
		}
	} else {
		if _, err := b.Client.Rest().UpdateInteractionResponse(i.ApplicationID(), i.Token(), responses.UpdateErrorComponentsf(p, "modules.music.commands.play.added.to.queue", []any{len(tracks)}, getMusicControllerComponents(nil))); err != nil {
			b.Logger.Error("Error while updating interaction message: ", err)
		}
	}
	if len(tracks) > 0 {
		player.Queue.Push(tracks...)
	}
}

func giveSearchSelection(b *dbot.Bot, p *message.Printer, e *events.ApplicationCommandInteractionCreate, tracks []lavalink.AudioTrack) {
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
	if _, err := e.Client().Rest().UpdateInteractionResponse(e.ApplicationID(), e.Token(), responses.UpdateErrorComponentsf(p, "modules.music.commands.play.search.selection", []any{len(tracks)})); err != nil {
		b.Logger.Error("Error while updating interaction message: ", err)
	}

	if _, err := e.Client().Rest().UpdateInteractionResponse(e.ApplicationID(), e.Token(),
		responses.UpdateSuccessComponentsf(p, "modules.music.autocomplete.select.songs", nil, discord.NewActionRow(
			discord.NewSelectMenu("play:search:"+e.ID().String(), p.Sprintf("modules.music.commands.play.select.songs"), options...).WithMaxValues(len(options)),
		)),
	); err != nil {
		b.Logger.Error("Error while updating interaction message: ", err)
	}

	go func() {
		collectorChan, cancel := bot.NewEventCollector(e.Client(), func(e *events.ComponentInteractionCreate) bool {
			return e.Data.CustomID() == "play:search:"+e.ID().String()
		})
		defer cancel()
		for {
			select {
			case e := <-collectorChan:
				if voiceState, ok := e.Client().Caches().VoiceStates().Get(*e.GuildID(), e.User().ID); !ok || voiceState.ChannelID == nil {
					if err := e.CreateMessage(responses.CreateErrorf(p, "modules.music.not.in.voice")); err != nil {
						b.Logger.Error("Failed to update error message: ", err)
					}
					continue
				}
				if err := e.DeferUpdateMessage(); err != nil {
					b.Logger.Error(err)
					return
				}
				var playTracks []lavalink.AudioTrack
				for _, value := range e.SelectMenuInteractionData().Values {
					index, _ := strconv.Atoi(value)
					playTracks = append(playTracks, tracks[index])
				}
				playAndQueue(b, p, e.BaseInteraction, playTracks...)
				return

			case <-time.After(time.Second * 30):
				if _, err := e.Client().Rest().UpdateInteractionResponse(e.ApplicationID(), e.Token(), responses.UpdateErrorComponentsf(p, "modules.music.commands.play.search.timed.out", nil)); err != nil {
					b.Logger.Error("Error while updating interaction message: ", err)
				}
				return
			}
		}
	}()
}
