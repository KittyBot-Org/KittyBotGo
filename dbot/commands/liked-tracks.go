package commands

import (
	"fmt"

	"github.com/KittyBot-Org/KittyBotGo/dbot"
	"github.com/KittyBot-Org/KittyBotGo/dbot/responses"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/handler"
	"github.com/disgoorg/utils/paginator"
	"github.com/lithammer/fuzzysearch/fuzzy"
)

func LikedTracks(b *dbot.Bot) handler.Command {
	return handler.Command{
		Create: discord.SlashCommandCreate{
			Name:        "liked-tracks",
			Description: "Lists/Removes a liked track.",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionSubCommand{
					Name:        "list",
					Description: "Lists all your liked tracks.",
				},
				discord.ApplicationCommandOptionSubCommand{
					Name:        "remove",
					Description: "Removes a liked track.",
					Options: []discord.ApplicationCommandOption{
						discord.ApplicationCommandOptionString{
							Name:         "track",
							Description:  "The track to remove",
							Required:     true,
							Autocomplete: true,
						},
					},
				},
				discord.ApplicationCommandOptionSubCommand{
					Name:        "clear",
					Description: "Clears all your liked tracks.",
				},
				/*discord.ApplicationCommandOptionSubCommand{
					Name:        "play",
					Description: "Plays a liked track.",
					Options: []discord.ApplicationCommandOption{
						discord.ApplicationCommandOptionString{
							Name:         "track",
							Description:  "The track to play",
							Required:     false,
							Autocomplete: true,
						},
					},
				},*/
			},
		},
		CommandHandlers: map[string]handler.CommandHandler{
			"list":   likedTracksListHandler(b),
			"remove": likedTracksRemoveHandler(b),
			"clear":  likedTracksClearHandler(b),
			"play":   likedTracksPlayHandler(b),
		},
		AutocompleteHandlers: map[string]handler.AutocompleteHandler{
			"remove": likedTrackAutocompleteHandler(b),
			//"play":   likedTrackAutocompleteHandler(b),
		},
	}
}

func likedTracksListHandler(b *dbot.Bot) handler.CommandHandler {
	return func(e *events.ApplicationCommandInteractionCreate) error {
		tracks, err := b.DB.LikedTracks().GetAll(e.User().ID)
		if err != nil {
			return err
		}
		if len(tracks) == 0 {
			return e.CreateMessage(responses.CreateErrorf("You haven't liked any tracks yet."))
		}
		var (
			pages         []string
			page          string
			tracksCounter int
		)
		for i, track := range tracks {
			trackStr := fmt.Sprintf("%d. [`%s`](<%s>)\n", i+1, track.Title, track.Query)
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

		return b.Paginator.Create(e.Respond, &paginator.Paginator{
			PageFunc: func(page int, embed *discord.EmbedBuilder) {
				embed.SetTitlef("You have `%d` liked tracks:", len(tracks)).SetDescription(pages[page])
			},
			MaxPages:        len(pages),
			ExpiryLastUsage: true,
			ID:              e.ID().String(),
		})
	}
}

func likedTracksRemoveHandler(b *dbot.Bot) handler.CommandHandler {
	return func(e *events.ApplicationCommandInteractionCreate) error {
		trackName := e.SlashCommandInteractionData().String("track")

		if err := b.DB.LikedTracks().Delete(e.User().ID, trackName); err != nil {
			return e.CreateMessage(responses.CreateErrorf("Failed to remove track from liked tracks. Please try again."))
		}
		return e.CreateMessage(responses.CreateSuccessf("Removed `%s` from liked tracks.", trackName))
	}
}

func likedTracksClearHandler(b *dbot.Bot) handler.CommandHandler {
	return func(e *events.ApplicationCommandInteractionCreate) error {
		if err := b.DB.LikedTracks().DeleteAll(e.User().ID); err != nil {
			return e.CreateMessage(responses.CreateErrorf("Failed to clear liked tracks. Please try again."))
		}
		return e.CreateMessage(responses.CreateSuccessf("Cleared all liked tracks."))
	}
}

func likedTracksPlayHandler(b *dbot.Bot) handler.CommandHandler {
	return func(e *events.ApplicationCommandInteractionCreate) error {
		return nil
	}
}

func likedTrackAutocompleteHandler(b *dbot.Bot) handler.AutocompleteHandler {
	return func(e *events.AutocompleteInteractionCreate) error {
		track := e.Data.String("track")
		likedTracks, err := b.DB.LikedTracks().GetAll(e.User().ID)
		if err != nil {
			return err
		}
		if (len(likedTracks) == 0) && track == "" {
			return e.Result(nil)
		}
		labels := make([]string, len(likedTracks))
		unsortedResult := make(map[string]string, len(likedTracks))
		i := 0
		for _, entry := range likedTracks {
			labels[i] = entry.Title
			unsortedResult[entry.Title] = entry.Title
			i++
		}

		if track == "" {
			var choices []discord.AutocompleteChoice
			for key, value := range unsortedResult {
				choices = append(choices, discord.AutocompleteChoiceString{
					Name:  key,
					Value: value,
				})
			}
			return e.Result(choices)
		}

		ranks := fuzzy.RankFindFold(track, labels)
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
}
