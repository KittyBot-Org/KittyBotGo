package commands

import (
	"fmt"

	"github.com/KittyBot-Org/KittyBotGo/dbot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/handler"
	"github.com/disgoorg/utils/paginator"
)

func Queue(b *dbot.Bot) handler.Command {
	return handler.Command{
		Create: discord.SlashCommandCreate{
			Name:        "queue",
			Description: "Shows the current track queue.",
		},
		Check: dbot.HasMusicPlayer(b).And(dbot.HasQueueItems(b)),
		CommandHandlers: map[string]handler.CommandHandler{
			"": queueHandler(b),
		},
	}
}

func queueHandler(b *dbot.Bot) handler.CommandHandler {
	return func(e *events.ApplicationCommandInteractionCreate) error {
		tracks := b.MusicPlayers.Get(*e.GuildID()).Queue.Tracks()

		var (
			pages         []string
			page          string
			tracksCounter int
		)
		for i, track := range tracks {
			trackStr := fmt.Sprintf("%d. %s - %s [%s]\n", i+1, formatTrack(track), track.Info().Length, discord.UserMention(track.UserData().(dbot.AudioTrackData).Requester))
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
				embed.SetTitlef("Queue(`%d`):", len(tracks)).SetDescription(pages[page])
			},
			MaxPages:        len(pages),
			ExpiryLastUsage: true,
			ID:              e.ID().String(),
		})
	}
}
