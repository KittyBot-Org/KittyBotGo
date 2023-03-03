package commands

import (
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/handler"
	"github.com/disgoorg/utils/paginator"

	"github.com/KittyBot-Org/KittyBotGo/dbot"
)

func History(bot *dbot.Bot) handler.Command {
	return handler.Command{
		Create: discord.SlashCommandCreate{
			Name:        "history",
			Description: "Shows the current history.",
		},
		Check: dbot.HasMusicPlayer(bot).And(dbot.HasHistoryItems(bot)),
		CommandHandlers: map[string]handler.CommandHandler{
			"": historyHandler(bot),
		},
	}
}

func historyHandler(b *dbot.Bot) handler.CommandHandler {
	return func(e *events.ApplicationCommandInteractionCreate) error {
		tracks := b.MusicPlayers.Get(*e.GuildID()).History.Tracks()
		var (
			pages         []string
			page          string
			tracksCounter int
		)
		for i := len(tracks) - 1; i >= 0; i-- {
			track := tracks[i]
			trackStr := fmt.Sprintf("%d. %s - %s [%s]\n", len(tracks)-i, formatTrack(track), track.Info().Length, discord.UserMention(track.UserData().(dbot.AudioTrackData).Requester))
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
				embed.SetTitlef("Currently `%d` tracks are in the history:", len(tracks)).SetDescription(pages[page])
			},
			MaxPages:        len(pages),
			ExpiryLastUsage: true,
			ID:              e.ID().String(),
		})
	}
}
