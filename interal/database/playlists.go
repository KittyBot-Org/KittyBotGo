package database

import (
	"github.com/disgoorg/disgolink/v2/lavalink"
	"github.com/disgoorg/snowflake/v2"
)

type Playlist struct {
	ID     int          `db:"id"`
	Name   string       `db:"name"`
	UserID snowflake.ID `db:"user_id"`
}

type PlaylistTrack struct {
	ID         int            `db:"id"`
	PlaylistID int            `db:"playlist_id"`
	Position   int            `db:"position"`
	Track      lavalink.Track `db:"track"`
}

func (d *Database) GetPlaylists(userID snowflake.ID) ([]Playlist, error) {
	var playlists []Playlist
	err := d.dbx.Select(&playlists, "SELECT * FROM playlists WHERE user_id = $1", userID)
	return playlists, err
}

func (d *Database) GetPlaylist(playlistID int) (Playlist, []PlaylistTrack, error) {
	var playlist Playlist
	err := d.dbx.Get(&playlist, "SELECT * FROM playlists WHERE id = $1", playlistID)
	if err != nil {
		return playlist, nil, err
	}

	var tracks []PlaylistTrack
	err = d.dbx.Select(&tracks, "SELECT * FROM playlist_tracks WHERE playlist_id = $1", playlistID)
	return playlist, tracks, err
}

func (d *Database) CreatePlaylist(userID snowflake.ID, name string) (Playlist, error) {
	var playlist Playlist
	err := d.dbx.Get(&playlist, "INSERT INTO playlists (name, user_id) VALUES ($1, $2) RETURNING *", name, userID)
	return playlist, err
}

func (d *Database) DeletePlaylist(playlistID int, userID snowflake.ID) error {
	_, err := d.dbx.Exec("DELETE FROM playlists WHERE id = $1 AND user_id = $2", playlistID, userID)
	return err
}

func (d *Database) AddTracksToPlaylist(playlistID int, tracks []lavalink.Track) error {
	playlistTracks := make([]PlaylistTrack, len(tracks))
	for i, track := range tracks {
		playlistTracks[i] = PlaylistTrack{
			PlaylistID: playlistID,
			Track:      track,
		}
	}
	_, err := d.dbx.NamedExec("INSERT INTO playlist_tracks (playlist_id, track) VALUES (:playlist_id, :track)", tracks)
	return err
}

func (d *Database) RemoveTracksFromPlaylist(trackIDs []int) error {
	_, err := d.dbx.NamedExec("DELETE FROM playlist_tracks WHERE id = :id", trackIDs)
	return err
}
