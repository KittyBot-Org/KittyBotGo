package handlers

import (
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"

	"github.com/KittyBot-Org/KittyBotGo/service/bot/res"
)

var historyCommand = discord.SlashCommandCreate{
	Name:        "history",
	Description: "Shows the history of the current guild",
	Options: []discord.ApplicationCommandOption{
		discord.ApplicationCommandOptionSubCommand{
			Name:        "clear",
			Description: "Clears the history",
		},
		discord.ApplicationCommandOptionSubCommand{
			Name:        "show",
			Description: "Shows the history",
		},
	},
}

func (h *Handlers) OnHistoryClear(e *handler.CommandEvent) error {
	err := h.Database.ClearHistory(*e.GuildID())
	if err != nil {
		return e.CreateMessage(res.CreateErr("Failed to clear history", err))
	}

	return e.CreateMessage(res.Createf("Cleared history"))
}

func (h *Handlers) OnHistoryShow(e *handler.CommandEvent) error {
	tracks, err := h.Database.GetHistory(*e.GuildID())
	if err != nil {
		return e.CreateMessage(res.CreateErr("Failed to get history", err))
	}

	if len(tracks) == 0 {
		return e.CreateMessage(res.Create("The history is empty"))
	}

	content := fmt.Sprintf("History(`%d`):\n", len(tracks))
	for i, track := range tracks {
		line := fmt.Sprintf("%d. %s\n", i+1, res.FormatTrack(track.Track, 0))
		if len([]rune(content))+len([]rune(line)) > 2000 {
			break
		}
		content += line
	}

	return e.CreateMessage(res.Create(content))
}
