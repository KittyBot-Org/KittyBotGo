package database

import (
	"database/sql"
	"errors"

	"github.com/disgoorg/snowflake/v2"
)

type QueueType int

const (
	QueueTypeDefault QueueType = iota
	QueueTypeLoopTrack
	QueueTypeLoopQueue
)

type Player struct {
	GuildID   snowflake.ID `db:"guild_id"`
	Node      string       `db:"node"`
	QueueType int          `db:"queue_type"`
}

func (d *Database) HasPlayer(guildID snowflake.ID) (bool, error) {
	var count int
	err := d.dbx.Get(&count, "SELECT COUNT(*) FROM players WHERE guild_id = $1", guildID)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (d *Database) GetPlayer(guildID snowflake.ID, node string) (*Player, []Track, error) {
	var player Player
	err := d.dbx.Get(&player, "SELECT * FROM players WHERE guild_id = $1", guildID)
	if errors.Is(err, sql.ErrNoRows) {
		_, err = d.dbx.Exec("INSERT INTO players (guild_id, node, queue_type) VALUES ($1, $2, $3)", guildID, node, QueueTypeDefault)
	}
	if err != nil {
		return nil, nil, err
	}

	var queue []Track
	err = d.dbx.Select(queue, "SELECT * FROM queue WHERE guild_id = $1 ORDER BY id DESC", guildID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, nil, err
	}

	for i := range queue {
		if err = queue[i].Unmarshal(); err != nil {
			return nil, nil, err
		}
	}

	return &player, queue, err
}

func (d *Database) UpdatePlayer(player Player) error {
	_, err := d.dbx.Exec("UPDATE players SET queue_type = $1, node = $2 WHERE guild_id = $3", player.QueueType, player.Node, player.GuildID)
	return err
}

func (d *Database) DeletePlayer(guildID snowflake.ID) error {
	_, err := d.dbx.Exec("DELETE FROM players WHERE guild_id = $1", guildID)
	return err
}
