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

func (c *Cmds) OnQueue(e *handler.CommandEvent) error {
	tracks, err := c.Database.GetQueue(*e.GuildID())
	if err != nil {
		return e.CreateMessage(res.CreateErr("Failed to get queue", err))
	}

	if len(tracks) == 0 {
		return e.CreateMessage(res.Create("The queue is empty"))
	}

	var content string
	for i, track := range tracks {
		line := fmt.Sprintf("%d. %s\n", i+1, res.FormatTrack(track.Track, 0))
		if len([]rune(content))+len([]rune(line)) > 2000 {
			break
		}
		content += line
	}

	return e.CreateMessage(res.Create(content))
}
