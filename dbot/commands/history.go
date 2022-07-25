package commands

import (
	"fmt"

	"github.com/KittyBot-Org/KittyBotGo/dbot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/utils/paginator"
	"golang.org/x/text/message"
)

var History = handler.Command{
	Create: discord.SlashCommandCreate{
		CommandName: "history",
		Description: "Shows the current history.",
	},
	Checks: dbot.HasMusicPlayer.And(dbot.HasHistoryItems),
	CommandHandler: map[string]handler.CommandHandler{
		"": historyHandler,
	},
}

func historyHandler(b *dbot.Bot, p *message.Printer, e *events.ApplicationCommandInteractionCreate) error {
	tracks := b.MusicPlayers.Get(*e.GuildID()).History.Tracks()

	var (
		pages         []string
		page          string
		tracksCounter int
	)
	for i := len(tracks) - 1; i >= 0; i-- {
		track := tracks[i]
		trackStr := fmt.Sprintf("%d. %s - %s [<@%s>]\n", len(tracks)-i, formatTrack(track), track.Info().Length, track.UserData().(dbot.AudioTrackData).Requester)
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
			embed.SetTitlef(p.Sprintf("modules.music.commands.history.title", len(tracks))).SetDescription(pages[page])
		},
		MaxPages:        len(pages),
		ExpiryLastUsage: true,
		ID:              e.ID().String(),
	})
}
