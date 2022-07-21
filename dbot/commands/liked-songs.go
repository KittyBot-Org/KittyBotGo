package commands

import (
	"fmt"

	"github.com/KittyBot-Org/KittyBotGo/dbot"
	"github.com/KittyBot-Org/KittyBotGo/dbot/responses"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/utils/paginator"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"golang.org/x/text/message"
)

var LikedSongs = dbot.Command{
	Create: discord.SlashCommandCreate{
		CommandName: "liked-songs",
		Description: "Lists/Removes/Plays a liked song.",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionSubCommand{
				CommandName: "list",
				Description: "Lists all your liked songs.",
			},
			discord.ApplicationCommandOptionSubCommand{
				CommandName: "remove",
				Description: "Removes a liked song.",
				Options: []discord.ApplicationCommandOption{
					discord.ApplicationCommandOptionString{
						OptionName:   "song",
						Description:  "The song to remove",
						Required:     true,
						Autocomplete: true,
					},
				},
			},
			discord.ApplicationCommandOptionSubCommand{
				CommandName: "clear",
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
	CommandHandler: map[string]dbot.CommandHandler{
		"list":   likedSongsListHandler,
		"remove": likedSongsRemoveHandler,
		"clear":  likedSongsClearHandler,
		"play":   likedSongsPlayHandler,
	},
	AutoCompleteHandler: map[string]dbot.AutocompleteHandler{
		"remove": likedSongAutocompleteHandler,
		//"play":   likedSongAutocompleteHandler,
	},
}

func likedSongsListHandler(b *dbot.Bot, p *message.Printer, e *events.ApplicationCommandInteractionCreate) error {
	tracks, err := b.DB.LikedSongs().GetAll(e.User().ID)
	if err != nil {
		return err
	}
	if len(tracks) == 0 {
		return e.CreateMessage(responses.CreateErrorf(p, "modules.music.commands.liked.songs.list.empty"))
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
			embed.SetTitlef(p.Sprintf("modules.music.commands.liked.songs.list.title", len(tracks))).SetDescription(pages[page])
		},
		MaxPages:        len(pages),
		ExpiryLastUsage: true,
		ID:              e.ID().String(),
	})
}

func likedSongsRemoveHandler(b *dbot.Bot, p *message.Printer, e *events.ApplicationCommandInteractionCreate) error {
	songName := e.SlashCommandInteractionData().String("song")

	if err := b.DB.LikedSongs().Delete(e.User().ID, songName); err != nil {
		return e.CreateMessage(responses.CreateErrorf(p, "modules.music.commands.liked.songs.remove.error"))
	}
	return e.CreateMessage(responses.CreateSuccessf(p, "modules.music.commands.liked.songs.remove.success", songName))
}

func likedSongsClearHandler(b *dbot.Bot, p *message.Printer, e *events.ApplicationCommandInteractionCreate) error {
	if err := b.DB.LikedSongs().DeleteAll(e.User().ID); err != nil {
		return e.CreateMessage(responses.CreateErrorf(p, "modules.music.commands.liked.songs.clear.error"))
	}
	return e.CreateMessage(responses.CreateSuccessf(p, "modules.music.commands.liked.songs.clear.success"))
}

func likedSongsPlayHandler(b *dbot.Bot, p *message.Printer, e *events.ApplicationCommandInteractionCreate) error {
	return nil
}

func likedSongAutocompleteHandler(b *dbot.Bot, _ *message.Printer, e *events.AutocompleteInteractionCreate) error {
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
