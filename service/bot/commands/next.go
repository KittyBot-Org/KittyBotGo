package commands

import (
	"context"
	"database/sql"
	"errors"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/disgolink/v2/lavalink"

	"github.com/KittyBot-Org/KittyBotGo/service/bot/res"
)

var next = discord.SlashCommandCreate{
	Name:        "next",
	Description: "Plays the next song in the queue",
}

func (c *Cmds) OnNext(e *handler.CommandEvent) error {
	player := c.Lavalink.Player(*e.GuildID())
	track, err := c.Database.NextTrack(*e.GuildID())
	if errors.Is(err, sql.ErrNoRows) {
		return e.CreateMessage(res.CreateError("No more songs in queue"))
	}
	if err != nil {
		return e.CreateMessage(res.CreateErr("Failed to get next song", err))
	}

	if err = player.Update(context.Background(), lavalink.WithTrack(track.Track)); err != nil {
		return e.CreateMessage(res.CreateErr("Failed to play next song", err))
	}
	return e.CreateMessage(res.Createf("Playing: %s", res.FormatTrack(track.Track, 0)))
}
