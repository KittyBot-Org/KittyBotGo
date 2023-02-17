package database

import (
	"database/sql"
	"errors"

	"github.com/disgoorg/snowflake/v2"
)

type QueueType int

const (
	QueueTypeNormal QueueType = iota
	QueueTypeRepeatTrack
	QueueTypeRepeatQueue
)

func (q QueueType) String() string {
	switch q {
	case QueueTypeNormal:
		return "Normal"
	case QueueTypeRepeatTrack:
		return "Repeat Track"
	case QueueTypeRepeatQueue:
		return "Repeat Queue"
	}
	return "Unknown"
}

type Player struct {
	GuildID   snowflake.ID `db:"guild_id"`
	Node      string       `db:"node"`
	QueueType QueueType    `db:"queue_type"`
}

func (d *DB) HasPlayer(guildID snowflake.ID) (bool, error) {
	var count int
	err := d.dbx.Get(&count, "SELECT COUNT(*) FROM players WHERE guild_id = $1", guildID)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (d *DB) GetPlayer(guildID snowflake.ID, node string) (*Player, error) {
	var player Player
	err := d.dbx.Get(&player, "SELECT * FROM players WHERE guild_id = $1", guildID)
	if errors.Is(err, sql.ErrNoRows) {
		_, err = d.dbx.Exec("INSERT INTO players (guild_id, node, queue_type) VALUES ($1, $2, $3)", guildID, node, QueueTypeNormal)
	}
	if err != nil {
		return nil, err
	}

	return &player, err
}

func (d *DB) UpdatePlayer(player Player) error {
	_, err := d.dbx.NamedExec("INSERT INTO players (guild_id, node, queue_type) VALUES (:guild_id, :node, :queue_type) ON CONFLICT (guild_id) DO UPDATE SET node = :node, queue_type = :queue_type", player)
	return err
}

func (d *DB) DeletePlayer(guildID snowflake.ID) error {
	_, err := d.dbx.Exec("DELETE FROM players WHERE guild_id = $1", guildID)
	return err
}
