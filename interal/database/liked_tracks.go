package database

import (
	"github.com/disgoorg/disgolink/v2/lavalink"
	"github.com/disgoorg/snowflake/v2"
)

type LikedTrack struct {
	ID     int            `db:"id"`
	UserID snowflake.ID   `db:"user_id"`
	Track  lavalink.Track `db:"track"`
}

func (d *Database) AddLikedTrack(userID snowflake.ID, track lavalink.Track) error {
	_, err := d.dbx.Exec("INSERT INTO liked_tracks (user_id, track) VALUES ($1, $2)", userID, track)
	return err
}

func (d *Database) RemoveLikedTrack(trackID int) error {
	_, err := d.dbx.Exec("DELETE FROM liked_tracks WHERE id = $1", trackID)
	return err
}

func (d *Database) GetLikedTracks(userID snowflake.ID) ([]LikedTrack, error) {
	var likedTracks []LikedTrack
	if err := d.dbx.Select(&likedTracks, "SELECT * FROM liked_tracks WHERE user_id = $1", userID); err != nil {
		return nil, err
	}

	return likedTracks, nil
}

func (d *Database) ClearLikedTracks(userID snowflake.ID) error {
	_, err := d.dbx.Exec("DELETE FROM liked_tracks WHERE user_id = $1", userID)
	return err
}

func (d *Database) GetLikedTrack(trackID int) (*LikedTrack, error) {
	var likedTrack LikedTrack
	err := d.dbx.Get(&likedTrack, "SELECT * FROM liked_tracks WHERE id = $1", trackID)
	if err != nil {
		return nil, err
	}

	return &likedTrack, nil
}

func (d *Database) FindLikedTrack(userID snowflake.ID, uri string) (*LikedTrack, error) {
	var likedTrack LikedTrack
	err := d.dbx.Get(&likedTrack, "SELECT * FROM liked_tracks WHERE user_id = $1 AND track -> 'info' ->> 'uri' = $2", userID, uri)
	if err != nil {
		return nil, err
	}

	return &likedTrack, nil
}
