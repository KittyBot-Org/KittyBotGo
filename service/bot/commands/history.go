package commands

import (
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/lavaqueue-plugin"

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

func (c *commands) OnHistoryClear(_ discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	if err := lavaqueue.ClearHistory(e.Ctx, c.Lavalink.Player(*e.GuildID()).Node(), *e.GuildID()); err != nil {
		return e.CreateMessage(res.CreateErr("Failed to clear history", err))
	}

	return e.CreateMessage(res.Createf("Cleared history"))
}

func (c *commands) OnHistoryShow(_ discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	tracks, err := lavaqueue.GetHistory(e.Ctx, c.Lavalink.Player(*e.GuildID()).Node(), *e.GuildID())
	if err != nil {
		return e.CreateMessage(res.CreateErr("Failed to get history", err))
	}

	if len(tracks) == 0 {
		return e.CreateMessage(res.Create("The history is empty"))
	}

	content := fmt.Sprintf("History(`%d`):\n", len(tracks))
	for i, track := range tracks {
		line := fmt.Sprintf("%d. %s\n", i+1, res.FormatTrack(track, 0))
		if len([]rune(content))+len([]rune(line)) > 2000 {
			break
		}
		content += line
	}

	return e.CreateMessage(res.Create(content))
}
