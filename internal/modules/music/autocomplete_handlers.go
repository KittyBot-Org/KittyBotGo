package music

import (
	"fmt"
	"strings"

	"github.com/KittyBot-Org/KittyBotGo/internal/kbot"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"golang.org/x/text/message"
)

func playAutocompleteHandler(b *kbot.Bot, _ *message.Printer, e *events.AutocompleteInteractionEvent) error {
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

func removeSongAutocompleteHandler(b *kbot.Bot, _ *message.Printer, e *events.AutocompleteInteractionEvent) error {
	player := b.MusicPlayers.Get(*e.GuildID())
	if player == nil || player.Queue.Len() == 0 {
		return e.Result(nil)
	}
	tracks := make([]string, player.Queue.Len())

	for i, track := range player.Queue.Tracks() {
		tracks[i] = fmt.Sprintf("%d. %s", i+1, track.Info().Title)
	}

	ranks := fuzzy.RankFindFold(e.Data.String("song"), tracks)

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

func likedSongAutocompleteHandler(b *kbot.Bot, _ *message.Printer, e *events.AutocompleteInteractionEvent) error {
	song := e.Data.String("song")
	likedSongs, err := b.DB.LikedSongs().GetAll(e.User().ID)
	if err != nil {
		return err
	}
	if (len(likedSongs) == 0) && song == "" {
		return e.Result(nil)
	}
	labels := make([]string, len(likedSongs))
	unsortedResult := make(map[string]string, len(likedSongs))
	i := 0
	for _, entry := range likedSongs {
		labels[i] = entry.Title
		unsortedResult[entry.Title] = entry.Title
		i++
	}

	if song == "" {
		var choices []discord.AutocompleteChoice
		for key, value := range unsortedResult {
			choices = append(choices, discord.AutocompleteChoiceString{
				Name:  key,
				Value: value,
			})
		}
		return e.Result(choices)
	}

	ranks := fuzzy.RankFindFold(song, labels)
	resultLen := len(ranks)
	if resultLen > 25 {
		resultLen = 25
	}
	result := make([]discord.AutocompleteChoice, resultLen+1)
	for ii, rank := range ranks {
		if ii >= resultLen {
			break
		}
		result[ii] = discord.AutocompleteChoiceString{
			Name:  rank.Target,
			Value: unsortedResult[rank.Target],
		}
	}
	return e.Result(result)
}
