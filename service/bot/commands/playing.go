package commands

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"

	"github.com/KittyBot-Org/KittyBotGo/service/bot/res"
)

var playing = discord.SlashCommandCreate{
	Name:        "playing",
	Description: "Shows the current playing song",
}

func (h *Cmds) OnPlaying(e *handler.CommandEvent) error {
	player := h.Player(*e.GuildID())
	if player.Track() == nil {
		return e.CreateMessage(res.CreateError("No song is currently playing"))
	}

	return e.CreateMessage(res.Createf("Playing: %s", res.FormatTrack(*player.Track(), player.Position())))
}
