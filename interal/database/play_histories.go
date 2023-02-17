package database

import (
	"time"

	"github.com/disgoorg/disgolink/v2/lavalink"
	"github.com/disgoorg/snowflake/v2"
)

type PlayHistory struct {
	ID       int            `db:"id"`
	UserID   snowflake.ID   `db:"user_id"`
	PlayedAt time.Time      `db:"played_at"`
	Track    lavalink.Track `db:"track"`
}

func (d *DB) AddPlayHistoryTracks(userID snowflake.ID, tracks []lavalink.Track) error {
	for _, track := range tracks {
		dbTrack := PlayHistory{
			UserID:   userID,
			PlayedAt: time.Now(),
			Track:    track,
		}
		if _, err := d.dbx.NamedExec("INSERT INTO play_histories (user_id, played_at, track) VALUES (:user_id, :played_at, :track) ON CONFLICT (user_id, track) DO UPDATE SET played_at = :played_at", dbTrack); err != nil {
			return err
		}
	}

	_, err := d.dbx.Exec("DELETE FROM play_histories WHERE user_id = $1 AND id NOT IN (SELECT id FROM play_histories WHERE user_id = $1 ORDER BY played_at DESC LIMIT 10)", userID)
	return err
}

func (d *DB) GetPlayHistoryTrack(trackID int) (*PlayHistory, error) {
	var history PlayHistory
	if err := d.dbx.Get(&history, "SELECT * FROM play_histories WHERE id = $1", trackID); err != nil {
		return nil, err
	}

	return &history, nil
}

func (d *DB) SearchPlayHistory(userID snowflake.ID, query string, limit int) ([]PlayHistory, error) {
	var history []PlayHistory
	if err := d.dbx.Select(&history, "SELECT * FROM queues WHERE usedr_id = $1 ORDER BY track -> 'info' ->> 'title' <->> $2 ASC LIMIT $3", userID, query, limit); err != nil {
		return nil, err
	}

	return history, nil
}
