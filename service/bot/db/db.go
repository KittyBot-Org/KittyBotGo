package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/disgoorg/snowflake/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

type Config struct {
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	Username string `toml:"username"`
	Password string `toml:"password"`
	Database string `toml:"database"`
	SSLMode  string `toml:"ssl_mode"`
}

func (c Config) String() string {
	return fmt.Sprintf("\n   Host: %s\n   Port: %dn   Username: %s\n   Password: %s\n   Database: %s\n   SSLMode: %s",
		c.Host,
		c.Port,
		c.Username,
		strings.Repeat("*", len(c.Password)),
		c.Database,
		c.SSLMode,
	)
}

func (c Config) PostgresDataSourceName() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host,
		c.Port,
		c.Username,
		c.Password,
		c.Database,
		c.SSLMode,
	)
}

func New(cfg Config, schema string) (*DB, error) {
	pgCfg, err := pgx.ParseConfig(cfg.PostgresDataSourceName())
	if err != nil {
		return nil, err
	}

	db, err := sqlx.Open("pgx", stdlib.RegisterConnConfig(pgCfg))
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err = db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	if _, err = db.ExecContext(ctx, schema); err != nil {
		return nil, fmt.Errorf("failed to execute schema: %w", err)
	}

	return &DB{
		dbx: db,
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
