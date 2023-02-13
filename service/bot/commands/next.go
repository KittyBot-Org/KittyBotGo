package commands

import (
	"context"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/disgolink/v2/lavalink"

	"github.com/KittyBot-Org/KittyBotGo/service/bot/res"
)

var next = discord.SlashCommandCreate{
	Name:        "next",
	Description: "Plays the next song in the queue",
}

func (h *Cmds) OnNext(e *handler.CommandEvent) error {
	player := h.Player(*e.GuildID())
	track, ok := player.Queue.Next()
	if !ok {
		return e.CreateMessage(res.CreateError("No more songs in queue"))
	}

	if err := player.Update(context.Background(), lavalink.WithTrack(track)); err != nil {
		return e.CreateMessage(res.CreateErr("Failed to play next song", err))
	}

	return e.CreateMessage(res.Createf("Playing: %s", res.FormatTrack(track, 0)))
}
