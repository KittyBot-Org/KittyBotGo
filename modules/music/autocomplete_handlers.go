package music

import (
	"context"
	"fmt"
	"github.com/DisgoOrg/disgo/core/events"
	"github.com/DisgoOrg/disgo/discord"
	"github.com/KittyBot-Org/KittyBotGo/internal/models"
	"github.com/KittyBot-Org/KittyBotGo/internal/types"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"golang.org/x/text/message"
	"strconv"
	"strings"
)

func playAutocompleteHandler(b *types.Bot, _ *message.Printer, e *events.AutocompleteInteractionEvent) error {
	var query string
	if q := e.Data.Options.String("query"); q != nil {
		query = *q
	}
	cache1 := b.PlayHistoryCache.Get(e.User.ID)
	var cache2 []models.LikedSong
	if err := b.DB.NewSelect().Model(&cache2).Where("user_id = ?", e.User.ID).Scan(context.TODO()); err != nil {
		b.Logger.Error("Failed to get music history entries: ", err)
		return err
	}
	if (len(cache1)+len(cache2) == 0) && query == "" {
		return e.Result(nil)
	}

	labels := make([]string, len(cache1)+len(cache2))
	unsortedResult := make(map[string]string, len(cache1)+len(cache2))
	i := 0
	for _, entry := range cache1 {
		title := "🔁" + entry.Title
		unsortedResult[title] = entry.Query
		labels[i] = title
		i++
	}

	for _, entry := range cache2 {
		title := "❤️" + entry.Title
		unsortedResult[title] = entry.Query
		labels[i] = title
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
			Name:  rank.Target,
			Value: unsortedResult[rank.Target],
		}
	}

	return e.Result(result)
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

func likedSongAutocompleteHandler(b *types.Bot, _ *message.Printer, e *events.AutocompleteInteractionEvent) error {
	song := *e.Data.Options.String("song")
	var likedSongs []models.LikedSong
	if err := b.DB.NewSelect().Model(&likedSongs).Where("user_id = ?", e.User.ID).Scan(context.TODO()); err != nil {
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
		unsortedResult[entry.Title] = strconv.Itoa(entry.ID)
		i++
	}

	if song == "" {
		return e.ResultMapString(unsortedResult)
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
