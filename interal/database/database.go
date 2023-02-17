package database

import (
	"context"
	_ "embed"
	"fmt"
	"strings"

	"github.com/disgoorg/snowflake/v2"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

//go:embed schema.sql
var schema string

type Config struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Database string `json:"database"`
	SSLMode  string `json:"ssl_mode"`
}

func (c Config) String() string {
	return fmt.Sprintf("Host: %s,\n  Port: %d,\n  Username: %s,\n  Password: %s,\n  Database: %s,\n  SSLMode: %s", c.Host, c.Port, c.Username, strings.Repeat("*", len(c.Password)), c.Database, c.SSLMode)
}

func New(ctx context.Context, cfg Config) (*DB, error) {
	dbx, err := sqlx.ConnectContext(ctx, "pgx", fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s", cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.Database, cfg.SSLMode))
	if err != nil {
		return nil, err
	}

	// execute schema
	if _, err = dbx.ExecContext(ctx, schema); err != nil {
		return nil, err
	}

	return &DB{
		dbx: dbx,
	}, nil
}

type DB struct {
	dbx *sqlx.DB
}

func (d *DB) Close() error {
	return d.dbx.Close()
}

type PlayTrack struct {
	ID   int    `db:"id"`
	Type string `db:"type"`
	Name string `db:"name"`
}

func (d *DB) SearchPlay(userID snowflake.ID, query string, limit int) ([]PlayTrack, error) {
	var tracks []PlayTrack
	if err := d.dbx.Select(&tracks, `SELECT * FROM(
		SELECT id, 'liked_track' as type, track -> 'info' ->> 'title' as name FROM liked_tracks WHERE user_id = $1
		UNION ALL
		SELECT id, 'playlist' as type, name FROM playlists WHERE user_id = $1
		UNION ALL
		SELECT id, 'play_history' as type, track -> 'info' ->> 'title' as name FROM play_histories  WHERE user_id = $1
		) t ORDER BY name <->> $2 ASC LIMIT $3;`, userID, query, limit); err != nil {
		return nil, err
	}

	return tracks, nil
}
