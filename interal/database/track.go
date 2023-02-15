package database

import (
	"github.com/disgoorg/disgolink/v2/lavalink"
	"github.com/disgoorg/snowflake/v2"
)

type Track struct {
	ID       int            `db:"id"`
	GuildID  snowflake.ID   `db:"guild_id"`
	Position int            `db:"position"`
	Track    lavalink.Track `db:"track"`
}
