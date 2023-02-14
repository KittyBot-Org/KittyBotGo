package database

import (
	"context"
	_ "embed"
	"fmt"
	"strings"

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

func New(ctx context.Context, cfg Config) (*Database, error) {
	dbx, err := sqlx.ConnectContext(ctx, "pgx", fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s", cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.Database, cfg.SSLMode))
	if err != nil {
		return nil, err
	}

	// execute schema
	if _, err = dbx.ExecContext(ctx, schema); err != nil {
		return nil, err
	}

	return &Database{
		dbx: dbx,
	}, nil
}

type Database struct {
	dbx *sqlx.DB
}

func (d *Database) Close() error {
	return d.dbx.Close()
}
