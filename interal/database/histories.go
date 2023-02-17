package database

import (
	"github.com/disgoorg/disgolink/v2/lavalink"
	"github.com/disgoorg/snowflake/v2"
)

func (d *DB) GetHistory(guildID snowflake.ID) ([]Track, error) {
	var history []Track
	if err := d.dbx.Select(&history, "SELECT * FROM histories WHERE guild_id = $1 ORDER BY position DESC", guildID); err != nil {
		return nil, err
	}

	return history, nil
}

func (d *DB) AddHistoryTracks(guildID snowflake.ID, tracks []lavalink.Track) error {
	dbTracks := make([]Track, len(tracks))
	for i, track := range tracks {
		dbTracks[i] = Track{
			GuildID: guildID,
			Track:   track,
		}
	}

	_, err := d.dbx.NamedExec("INSERT INTO histories (guild_id, track) VALUES (:guild_id, :track)", dbTracks)
	return err
}

func (d *DB) PreviousHistoryTrack(guildID snowflake.ID) (*Track, error) {
	var track Track
	err := d.dbx.Get(&track, "DELETE FROM histories WHERE position = (SELECT MAX(position) from histories WHERE guild_id = $1) RETURNING *", guildID)
	if err != nil {
		return nil, err
	}

	return &track, nil
}

func (d *DB) RemoveHistoryTrack(trackID int) error {
	_, err := d.dbx.Exec("DELETE FROM histories WHERE id = $1", trackID)
	return err
}

func (d *DB) ClearHistory(guildID snowflake.ID) error {
	_, err := d.dbx.Exec("DELETE FROM histories WHERE guild_id = $1", guildID)
	return err
}
