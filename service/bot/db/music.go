package db

import (
	"context"
	"encoding/json"
	"time"

	"github.com/disgoorg/disgolink/v4/lavalink"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/kittybot-org/kittybotgo/service/bot/db/dbsqlc"
)

func (d *DB) LoadPlaylist(ctx context.Context, playlistID int64) (dbsqlc.GetPlaylistRow, []lavalink.Track, error) {
	playlist, err := d.GetPlaylist(ctx, playlistID)
	if err != nil {
		return dbsqlc.GetPlaylistRow{}, nil, err
	}
	rows, err := d.ListPlaylistTracks(ctx, playlistID)
	if err != nil {
		return dbsqlc.GetPlaylistRow{}, nil, err
	}
	tracks, err := decodeTracks(rows)
	if err != nil {
		return dbsqlc.GetPlaylistRow{}, nil, err
	}
	return playlist, tracks, nil
}

func (d *DB) AddTracksToPlaylist(ctx context.Context, playlistID int64, tracks []lavalink.Track) error {
	for _, track := range tracks {
		encoded, err := json.Marshal(track)
		if err != nil {
			return err
		}
		if err = d.InsertPlaylistTrack(ctx, dbsqlc.InsertPlaylistTrackParams{
			PlaylistID: playlistID,
			Track:      encoded,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (d *DB) AddPlayHistoryTracks(ctx context.Context, userID int64, tracks []lavalink.Track) error {
	tx, err := d.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	q := d.WithTx(tx)
	playedAt := pgtype.Timestamp{Time: time.Now(), Valid: true}
	for _, track := range tracks {
		encoded, err := json.Marshal(track)
		if err != nil {
			return err
		}
		if err = q.UpsertPlayHistory(ctx, dbsqlc.UpsertPlayHistoryParams{
			UserID:   userID,
			PlayedAt: playedAt,
			Track:    encoded,
		}); err != nil {
			return err
		}
	}
	if err = q.TrimPlayHistory(ctx, userID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func decodeTracks(rows []dbsqlc.PlaylistTrack) ([]lavalink.Track, error) {
	tracks := make([]lavalink.Track, len(rows))
	for i, row := range rows {
		if err := json.Unmarshal(row.Track, &tracks[i]); err != nil {
			return nil, err
		}
	}
	return tracks, nil
}
