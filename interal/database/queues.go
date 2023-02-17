package database

import (
	"github.com/disgoorg/disgolink/v2/lavalink"
	"github.com/disgoorg/snowflake/v2"
)

func (d *Database) GetQueue(guildID snowflake.ID) ([]Track, error) {
	var queue []Track
	if err := d.dbx.Select(&queue, "SELECT * FROM queues WHERE guild_id = $1 ORDER BY position ASC", guildID); err != nil {
		return nil, err
	}

	return queue, nil
}

func (d *Database) AddQueueTracks(guildID snowflake.ID, tracks []lavalink.Track) error {
	dbTracks := make([]Track, len(tracks))
	for i, track := range tracks {
		dbTracks[i] = Track{
			GuildID: guildID,
			Track:   track,
		}
	}

	_, err := d.dbx.NamedExec("INSERT INTO queues (guild_id, track) VALUES (:guild_id, :track)", dbTracks)
	return err
}

func (d *Database) NextQueueTrack(guildID snowflake.ID) (*Track, error) {
	var track Track
	err := d.dbx.Get(&track, "DELETE FROM queues WHERE position = (SELECT MIN(position) from queues WHERE guild_id = $1) RETURNING *", guildID)
	if err != nil {
		return nil, err
	}

	return &track, nil
}

func (d *Database) RemoveQueueTrack(trackID int) error {
	_, err := d.dbx.Exec("DELETE FROM queues WHERE id = $1", trackID)
	return err
}

func (d *Database) ClearQueue(guildID snowflake.ID) error {
	_, err := d.dbx.Exec("DELETE FROM queues WHERE guild_id = $1", guildID)
	return err
}

func (d *Database) ShuffleQueue(guildID snowflake.ID) error {
	var queueSize int
	err := d.dbx.Get(&queueSize, "SELECT COUNT(*) FROM queues WHERE guild_id = $1", guildID)
	if err != nil {
		return err
	}
	_, err = d.dbx.Exec("UPDATE queues SET position = floor(random() * $1) + 1 WHERE guild_id = $2", queueSize, guildID)
	return err
}
