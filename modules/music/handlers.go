package music

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/DisgoOrg/disgo/core"
	"github.com/DisgoOrg/disgo/core/events"
	"github.com/DisgoOrg/disgo/discord"
	"github.com/DisgoOrg/disgolink/lavalink"
	"github.com/DisgoOrg/snowflake"
	source_extensions "github.com/DisgoOrg/source-extensions-plugin"
	"github.com/DisgoOrg/utils/paginator"
	"github.com/KittyBot-Org/KittyBotGo/internal/models"
	"github.com/KittyBot-Org/KittyBotGo/internal/types"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"golang.org/x/text/message"
)

var (
	urlPattern    = regexp.MustCompile("^https?://[-a-zA-Z0-9+&@#/%?=~_|!:,.;]*[-a-zA-Z0-9+&@#/%=~_|]?")
	searchPattern = regexp.MustCompile(`^(.{2})search:(.+)`)

	bassBoost = &lavalink.Equalizer{
		0:  0.2,
		1:  0.15,
		2:  0.1,
		3:  0.05,
		4:  0.0,
		5:  -0.05,
		6:  -0.1,
		7:  -0.1,
		8:  -0.1,
		9:  -0.1,
		10: -0.1,
		11: -0.1,
		12: -0.1,
		13: -0.1,
		14: -0.1,
	}
)

func playHandler(b *types.Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) error {
	data := e.SlashCommandInteractionData()

	query := *data.Options.String("query")
	if searchProvider := data.Options.String("searchProvider"); searchProvider != nil {
		query = lavalink.SearchType(*searchProvider).Apply(query)
	} else {
		if !urlPattern.MatchString(query) && !searchPattern.MatchString(query) {
			query = lavalink.SearchTypeYoutube.Apply(query)
		}
	}

	if err := e.DeferCreateMessage(false); err != nil {
		return err
	}
	return b.Lavalink.BestRestClient().LoadItemHandler(context.TODO(), query, lavalink.NewResultHandler(
		func(track lavalink.AudioTrack) {
			b.PlayHistoryCache.Add(e.User.ID, query, track.Info().Title)
			playAndQueue(b, e.CreateInteraction, track)
		},
		func(playlist lavalink.AudioPlaylist) {
			b.PlayHistoryCache.Add(e.User.ID, query, playlist.Name())
			playAndQueue(b, e.CreateInteraction, playlist.Tracks()...)
		},
		func(tracks []lavalink.AudioTrack) {
			b.PlayHistoryCache.Add(e.User.ID, query, *data.Options.String("query"))
			giveSearchSelection(b, e, tracks)
		},
		func() {
			if _, err := e.UpdateOriginalMessage(discord.NewMessageUpdateBuilder().SetContent("No results found for your query.").Build()); err != nil {
				b.Logger.Error(err)
			}
		},
		func(ex lavalink.FriendlyException) {
			if _, err := e.UpdateOriginalMessage(discord.NewMessageUpdateBuilder().SetContent("There was an error with your request. Please try again\nError: " + ex.Message).Build()); err != nil {
				b.Logger.Error(err)
			}
		},
	))
}

func playAutocompleteHandler(b *types.Bot, p *message.Printer, e *events.AutocompleteInteractionEvent) error {
	var query string
	if q := e.Data.Options.String("query"); q != nil {
		query = *q
	}
	cache, ok := b.PlayHistoryCache.Get(e.User.ID)
	if (!ok || len(cache) == 0) && query == "" {
		return e.Result(nil)
	}

	labels := make([]string, len(cache))
	unsortedResult := make(map[string]string, len(cache))
	i := 0
	for _, entry := range cache {
		unsortedResult[entry.Title] = entry.Query
		i++
	}

	if query == "" {
		return e.ResultMapString(unsortedResult)
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
			Name:  "🔁" + rank.Target,
			Value: unsortedResult[rank.Target],
		}
	}

	return e.Result(result)
}

func playAndQueue(b *types.Bot, i core.CreateInteraction, tracks ...lavalink.AudioTrack) {
	player := b.MusicPlayers.Get(*i.GuildID)
	if player == nil {
		player = b.MusicPlayers.New(*i.GuildID, types.PlayerTypeMusic, types.LoopingTypeOff)
		b.MusicPlayers.Add(player)
	}
	var voiceChannelID snowflake.Snowflake
	if voiceState := i.Member.VoiceState(); voiceState == nil || voiceState.ChannelID == nil {
		if _, err := i.UpdateOriginalMessage(discord.NewMessageUpdateBuilder().SetContent("You need to be connected to a voice channel to play music").ClearContainerComponents().Build()); err != nil {
			b.Logger.Error(err)
		}
		return
	} else {
		voiceChannelID = *voiceState.ChannelID
	}
	if voiceState := i.Guild().SelfMember().VoiceState(); voiceState == nil || voiceState.ChannelID == nil || *voiceState.ChannelID != voiceChannelID {
		if err := b.Bot.AudioController.Connect(context.TODO(), *i.GuildID, voiceChannelID); err != nil {
			if _, err = i.UpdateOriginalMessage(discord.NewMessageUpdateBuilder().SetContent("I couldn't connect to your voice channel. Please make sure I have all required permissions!").ClearContainerComponents().Build()); err != nil {
				b.Logger.Error(err)
			}
			return
		}
	}

	for ii := range tracks {
		tracks[ii].SetUserData(models.AudioTrackData{
			Requester: i.User.ID,
		})
	}

	if player.PlayingTrack() == nil {
		track := tracks[0]
		if len(tracks) > 0 {
			tracks = tracks[1:]
		}
		if err := player.Play(track); err != nil {
			if _, err = i.UpdateOriginalMessage(discord.NewMessageUpdateBuilder().SetContent("There was an error with playing your song. Please try again").ClearContainerComponents().Build()); err != nil {
				b.Logger.Error(err)
			}
			return
		}
		if _, err := i.UpdateOriginalMessage(discord.NewMessageUpdateBuilder().SetContentf("Now playing: [`%s`](<%s>)", track.Info().Title, *track.Info().URI).ClearContainerComponents().Build()); err != nil {
			b.Logger.Error(err)
		}
	} else {
		if _, err := i.UpdateOriginalMessage(discord.NewMessageUpdateBuilder().SetContentf("Added %d tracks into queue", len(tracks)).ClearContainerComponents().Build()); err != nil {
			b.Logger.Error(err)
		}
	}
	if len(tracks) > 0 {
		player.Queue.Push(tracks...)
	}
}

func giveSearchSelection(b *types.Bot, event *events.ApplicationCommandInteractionEvent, tracks []lavalink.AudioTrack) {
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
	if _, err := event.UpdateOriginalMessage(discord.NewMessageUpdateBuilder().
		SetContent("Select songs to play").
		AddActionRow(discord.NewSelectMenu(discord.CustomID("play:search:"+event.ID), "Select songs to play", options...).WithMaxValues(len(options))).
		Build()); err != nil {
		b.Logger.Error(err)
	}
	go func() {
		coll, cancel := b.Bot.Collectors.NewComponentInteractionCollector(func(interaction *core.ComponentInteraction) bool {
			return interaction.Data.ID() == discord.CustomID("play:search:"+event.ID)
		})
		defer cancel()
		for {
			select {
			case i := <-coll:
				if voiceState := i.Member.VoiceState(); voiceState == nil || voiceState.ChannelID == nil {
					if err := i.CreateMessage(discord.NewMessageCreateBuilder().SetContent("You need to be connected to a voice channel to play music").SetEphemeral(true).Build()); err != nil {
						b.Logger.Error(err)
					}
					continue
				}
				if err := i.DeferUpdateMessage(); err != nil {
					b.Logger.Error(err)
					return
				}
				var playTracks []lavalink.AudioTrack
				for _, value := range i.SelectMenuInteractionData().Values {
					index, _ := strconv.Atoi(value)
					playTracks = append(playTracks, tracks[index])
				}
				playAndQueue(b, i.CreateInteraction, playTracks...)
				return

			case <-time.After(time.Second * 30):
				if _, err := event.UpdateOriginalMessage(discord.NewMessageUpdateBuilder().SetContent("Search timed out").ClearContainerComponents().Build()); err != nil {
					b.Logger.Error(err)
				}
				return
			}
		}
	}()
}

func queueHandler(b *types.Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) error {
	player := b.MusicPlayers.Get(*e.GuildID)
	if player == nil {
		return e.CreateMessage(discord.MessageCreate{Content: "No music player found"})
	}

	tracks := player.Queue.Tracks()
	if len(tracks) == 0 {
		return e.CreateMessage(discord.MessageCreate{Content: "The queue is empty"})
	}

	var (
		pages         []string
		page          string
		tracksCounter int
	)
	for i, track := range tracks {
		trackStr := fmt.Sprintf("%d. [`%s`](<%s>) - %s[<@%s>]\n", i+1, track.Info().Title, *track.Info().URI, track.Info().Length, track.UserData().(models.AudioTrackData).Requester)
		if len(page)+len(trackStr) > 4096 || tracksCounter >= 10 {
			pages = append(pages, page)
			page = ""
			tracksCounter = 0
		}
		page += trackStr
		tracksCounter++
	}
	if len(page) > 0 {
		pages = append(pages, page)
	}

	return b.Paginator.Create(e.CreateInteraction, &paginator.Paginator{
		PageFunc: func(page int, embed *discord.EmbedBuilder) discord.Embed {
			return embed.SetTitlef("Currently there are %d songs in the queue:", len(tracks)).SetDescription(pages[page]).Build()
		},
		MaxPages:        len(pages),
		Expiry:          time.Now(),
		ExpiryLastUsage: true,
	})
}

func removeSongHandler(b *types.Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) error {
	player := b.MusicPlayers.Get(*e.GuildID)
	strIndex := *e.SlashCommandInteractionData().Options.String("song")
	index, err := strconv.Atoi(strIndex)
	if err != nil {
		return e.CreateMessage(discord.NewMessageCreateBuilder().SetContent("Invalid song index").Build())
	}
	if player.Queue.Len() == 0 {
		return e.CreateMessage(discord.NewMessageCreateBuilder().SetContent("The queue is empty").Build())
	}

	removeTrack := player.Queue.Get(index - 1)
	if removeTrack == nil {
		return e.CreateMessage(discord.NewMessageCreateBuilder().SetContentf("No track found with index %d", index).Build())
	}

	player.Queue.Remove(index - 1)
	return e.CreateMessage(discord.NewMessageCreateBuilder().
		SetContentf("Removed song [`%s`](<%s>) at index `%d` from the queue", removeTrack.Info().Title, *removeTrack.Info().URI, index).
		Build(),
	)
}

func removeUserSongsHandler(b *types.Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) error {
	player := b.MusicPlayers.Get(*e.GuildID)
	userID := *e.SlashCommandInteractionData().Options.Snowflake("user")
	if player.Queue.Len() == 0 {
		return e.CreateMessage(discord.NewMessageCreateBuilder().SetContent("The queue is empty").Build())
	}

	removedTracks := 0
	for i, track := range player.Queue.Tracks() {
		if track.UserData().(models.AudioTrackData).Requester == userID {
			player.Queue.Remove(i - removedTracks)
			removedTracks++
		}
	}
	if removedTracks == 0 {
		return e.CreateMessage(discord.NewMessageCreateBuilder().
			SetContentf("No track from user <@%s> found", userID).
			SetAllowedMentions(&discord.AllowedMentions{}).
			Build(),
		)
	}

	return e.CreateMessage(discord.NewMessageCreateBuilder().
		SetContentf("Removed `%d` songs from <@%s> from the queue", removedTracks, userID).
		SetAllowedMentions(&discord.AllowedMentions{}).
		Build(),
	)
}

func removeSongAutocompleteHandler(b *types.Bot, p *message.Printer, e *events.AutocompleteInteractionEvent) error {
	player := b.MusicPlayers.Get(*e.GuildID)
	if player == nil || player.Queue.Len() == 0 {
		return e.ResultMapInt(nil)
	}
	tracks := make([]string, player.Queue.Len())

	for i, track := range player.Queue.Tracks() {
		tracks[i] = fmt.Sprintf("%d. %s", i+1, track.Info().Title)
	}

	ranks := fuzzy.RankFindFold(*e.Data.Options.String("song"), tracks)

	choicesLen := len(ranks)
	if choicesLen > 25 {
		choicesLen = 25
	}
	choices := make([]discord.AutocompleteChoice, choicesLen)

	for i, rank := range ranks {
		if i >= 25 {
			break
		}
		choices[i] = discord.AutocompleteChoiceString{
			Name:  rank.Target,
			Value: strings.SplitN(rank.Target, ".", 2)[0],
		}
	}
	return e.Result(choices)
}

func clearQueueHandler(b *types.Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) error {
	player := b.MusicPlayers.Get(*e.GuildID)
	if player.Queue.Len() == 0 {
		return e.CreateMessage(discord.NewMessageCreateBuilder().SetContent("The queue is already empty").Build())
	}

	player.Queue.Clear()
	return e.CreateMessage(discord.NewMessageCreateBuilder().SetContent("Cleared the queue").Build())
}

func stopHandler(b *types.Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) error {
	player := b.MusicPlayers.Get(*e.GuildID)
	if err := player.Destroy(); err != nil {
		err = e.CreateMessage(discord.NewMessageCreateBuilder().
			SetContent("Failed to stop player. Please try again").
			Build())
		return err
	}
	if err := b.Bot.AudioController.Disconnect(context.TODO(), *e.GuildID); err != nil {
		err = e.CreateMessage(discord.NewMessageCreateBuilder().
			SetContent("Failed to disconnect from voice channel. Please try again").
			Build())
		return err
	}
	b.MusicPlayers.Delete(*e.GuildID)
	return e.CreateMessage(discord.NewMessageCreateBuilder().
		SetContent("Stopped player").
		Build())
}

func loopHandler(b *types.Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) error {
	data := e.SlashCommandInteractionData()
	player := b.MusicPlayers.Get(*e.GuildID)
	loopingType := types.LoopingType(*data.Options.Int("looping-type"))
	player.Queue.SetType(loopingType)
	emoji := ""
	switch loopingType {
	case types.LoopingTypeRepeatSong:
		emoji = "🔂"
	case types.LoopingTypeRepeatQueue:
		emoji = "🔁"
	}
	return e.CreateMessage(discord.NewMessageCreateBuilder().
		SetContentf("%s Looping: %s", emoji, loopingType).
		Build())
}

func nowPlayingHandler(b *types.Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) error {
	player := b.MusicPlayers.Get(*e.GuildID)
	if player.PlayingTrack() == nil {
		return e.CreateMessage(discord.NewMessageCreateBuilder().
			SetContent("No track is currently playing").
			SetEphemeral(true).
			Build())
	}
	i := player.PlayingTrack().Info()
	embed := discord.NewEmbedBuilder().
		SetColor(0xe24f96).
		SetTitle("Currently Playing:").
		SetDescriptionf("[`%s`](<%s>)", i.Title, *i.URI).
		AddField("Author", i.Author, true).
		AddField("Requested by", "todo", true).
		AddField("Volume", fmt.Sprintf("%d%%", player.Volume()), true).
		SetImage(getArtworkURL(player.PlayingTrack())).
		SetFooter("Tracks in queue: "+strconv.Itoa(player.Queue.Len()), "")
	if !i.IsStream {
		bar := [10]string{"▬", "▬", "▬", "▬", "▬", "▬", "▬", "▬", "▬", "▬"}
		t1 := player.Position()
		t2 := i.Length
		p := int(float64(t1) / float64(t2) * 10)
		bar[p] = "🔘"
		loopString := ""
		if player.Queue.LoopingType() == types.LoopingTypeRepeatSong {
			loopString = "🔂"
		}
		if player.Queue.LoopingType() == types.LoopingTypeRepeatQueue {
			loopString = "🔁"
		}
		embed.Description += fmt.Sprintf("\n\n%s / %s %s\n%s", formatPosition(t1), formatPosition(t2), loopString, bar)
	}
	return e.CreateMessage(discord.NewMessageCreateBuilder().
		SetEmbeds(embed.Build()).
		Build())
}

func pauseHandler(b *types.Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) error {
	player := b.MusicPlayers.Get(*e.GuildID)
	pause := true
	pauseStr := "pause"
	pauseStr2 := "Paused"
	if player.Paused() {
		pause = false
		pauseStr = "resume"
		pauseStr2 = "Resumed"
	}
	if err := player.Pause(pause); err != nil {
		return e.CreateMessage(discord.NewMessageCreateBuilder().
			SetContentf("Failed to %s player. Please try again", pauseStr).
			Build())
	}
	return e.CreateMessage(discord.NewMessageCreateBuilder().
		SetContentf("⏯ %s player", pauseStr2).
		Build())
}

func volumeHandler(b *types.Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) error {
	player := b.MusicPlayers.Get(*e.GuildID)
	volume := *e.SlashCommandInteractionData().Options.Int("volume")
	if err := player.SetVolume(volume); err != nil {
		return e.CreateMessage(discord.NewMessageCreateBuilder().
			SetContent("Failed to set volume. Please try again").
			Build())
	}
	return e.CreateMessage(discord.NewMessageCreateBuilder().
		SetContentf("🔉 Volume set to %d", volume).
		Build())
}

func bassBoostHandler(b *types.Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) error {
	player := b.MusicPlayers.Get(*e.GuildID)
	enable := *e.SlashCommandInteractionData().Options.Bool("enable")
	enableStr := "enabled"
	if enable {
		if err := player.Filters().SetEqualizer(bassBoost).Commit(); err != nil {
			return e.CreateMessage(discord.NewMessageCreateBuilder().
				SetContent("Failed to enable bass boost. Please try again").
				Build())
		}
	} else {
		enableStr = "disabled"
		if err := player.Filters().SetEqualizer(&lavalink.Equalizer{}).Commit(); err != nil {
			return e.CreateMessage(discord.NewMessageCreateBuilder().
				SetContent("Failed to set volume. Please try again").
				Build())
		}
	}
	return e.CreateMessage(discord.NewMessageCreateBuilder().
		SetContentf("Bass boost %s", enableStr).
		Build())
}

func seekHandler(b *types.Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) error {
	data := e.SlashCommandInteractionData()
	player := b.MusicPlayers.Get(*e.GuildID)
	position := *data.Options.Int("position")
	timeUnit := lavalink.Second
	if timeUnitPtr := data.Options.Int("time-unit"); timeUnitPtr != nil {
		timeUnit = lavalink.Duration(*timeUnitPtr)
	}

	finalPosition := lavalink.Duration(position) * timeUnit
	if finalPosition > player.PlayingTrack().Info().Length {
		return e.CreateMessage(discord.NewMessageCreateBuilder().
			SetContent("Position is too big").
			Build())
	}
	if err := player.Seek(finalPosition); err != nil {
		return e.CreateMessage(discord.NewMessageCreateBuilder().
			SetContent("Failed to seek. Please try again").
			Build())
	}
	return e.CreateMessage(discord.NewMessageCreateBuilder().
		SetContentf("⏩ Seeked to %s", formatPosition(finalPosition)).
		Build())
}

func skipHandler(b *types.Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) error {
	player := b.MusicPlayers.Get(*e.GuildID)
	nextTrack := player.Queue.Pop()
	if nextTrack == nil {
		return e.CreateMessage(discord.NewMessageCreateBuilder().
			SetContent("No more tracks in queue").
			Build())
	}
	if err := player.Play(nextTrack); err != nil {
		return e.CreateMessage(discord.NewMessageCreateBuilder().
			SetContent("Failed to skip track. Please try again").
			Build())
	}
	return e.CreateMessage(discord.NewMessageCreateBuilder().
		SetContentf("⏭ Skipped track.\nNow playing: [`%s`](<%s>) - %s", nextTrack.Info().Title, *nextTrack.Info().URI, nextTrack.Info().Length).
		Build())
}

func shuffleHandler(b *types.Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) error {
	player := b.MusicPlayers.Get(*e.GuildID)
	if player.Queue.Len() == 0 {
		return e.CreateMessage(discord.NewMessageCreateBuilder().
			SetContent("No tracks in queue to shuffle").
			Build())
	}
	player.Queue.Shuffle()
	return e.CreateMessage(discord.NewMessageCreateBuilder().
		SetContent("🔀 Shuffled queue").
		Build())
}

func formatPosition(position lavalink.Duration) string {
	if position == 0 {
		return "0:00"
	}
	return fmt.Sprintf("%d:%02d", position.Minutes(), position.SecondsPart())
}

func getArtworkURL(track lavalink.AudioTrack) string {
	switch track.Info().SourceName {
	case "youtube":
		return "https://i.ytimg.com/vi/" + track.Info().Identifier + "/maxresdefault.jpg"
	case "twitch":
		return "https://static-cdn.jtvnw.net/previews-ttv/live_user_" + track.Info().Identifier + "-440x248.jpg"

	case "spotify", "applemusic":
		isrcTrack, ok := track.(*source_extensions.ISRCAudioTrack)
		if ok && isrcTrack.ArtworkURL != nil {
			return *isrcTrack.ArtworkURL
		}
	}
	return ""
}
