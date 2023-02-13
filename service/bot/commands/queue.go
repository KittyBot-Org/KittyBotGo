package commands

import (
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"

	"github.com/KittyBot-Org/KittyBotGo/service/bot/res"
)

var queue = discord.SlashCommandCreate{
	Name:        "queue",
	Description: "Shows the current queue",
}

func (h *Cmds) OnQueue(e *handler.CommandEvent) error {
	player := h.Player(*e.GuildID())
	if len(player.Queue.Tracks) == 0 {
		return e.CreateMessage(res.CreateError("No songs in queue"))
	}

	var content string
	for i, track := range player.Queue.Tracks {
		line := fmt.Sprintf("%d. %s\n", i+1, res.FormatTrack(track, 0))
		if len([]rune(content))+len([]rune(line)) > 2000 {
			break
		}
		content += line
	}

	return e.CreateMessage(res.Create(content))
}
