package commands

import (
	"database/sql"
	"errors"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"

	"github.com/KittyBot-Org/KittyBotGo/service/bot/res"
)

var shuffle = discord.SlashCommandCreate{
	Name:        "shuffle",
	Description: "Shuffles the queue",
}

func (c *Cmds) OnShuffle(e *handler.CommandEvent) error {
	err := c.Database.ShuffleQueue(*e.GuildID())
	if errors.Is(err, sql.ErrNoRows) {
		return e.CreateMessage(res.CreateError("No more songs in queue"))
	}
	if err != nil {
		return e.CreateMessage(res.CreateErr("Failed to get next song", err))
	}

	return e.CreateMessage(res.Createf("Shuffled queue"))
}
