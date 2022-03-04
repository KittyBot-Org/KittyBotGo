package music

import (
	"fmt"
	"github.com/DisgoOrg/disgo/core/events"
	"github.com/DisgoOrg/disgo/discord"
	"github.com/KittyBot-Org/KittyBotGo/internal/types"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"golang.org/x/text/message"
	"strings"
)

func playAutocompleteHandler(b *types.Bot, _ *message.Printer, e *events.AutocompleteInteractionEvent) error {
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
