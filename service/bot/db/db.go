package db

import (
	"context"
	"embed"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/topi314/gomigrate"
	"github.com/topi314/gomigrate/drivers/postgres"

	"github.com/kittybot-org/kittybotgo/service/bot/db/dbsqlc"
)

//go:embed migrations/*.sql
var migrations embed.FS

type Config struct {
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	Username string `toml:"username"`
	Password string `toml:"password"`
	Database string `toml:"database"`
	SSLMode  string `toml:"ssl_mode"`
}

func (c Config) String() string {
	return fmt.Sprintf("\n   Host: %s\n   Port: %d\n   Username: %s\n   Password: %s\n   Database: %s\n   SSLMode: %s",
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

func New(cfg Config) (*DB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	poolCfg, err := pgxpool.ParseConfig(cfg.PostgresDataSourceName())
	if err != nil {
		return nil, err
	}
	poolCfg.ConnConfig.Tracer = otelpgx.NewTracer()
	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	if err = pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	if err = migrate(ctx, cfg.PostgresDataSourceName()); err != nil {
		pool.Close()
		return nil, err
	}

	return &DB{
		pool:    pool,
		Queries: dbsqlc.New(pool),
	}, nil
}

func migrate(ctx context.Context, dsn string) error {
	cfg, err := pgx.ParseConfig(dsn)
	if err != nil {
		return err
	}
	// The initial migration is one script with several statements.
	cfg.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
	sqlDB := stdlib.OpenDB(*cfg)
	defer sqlDB.Close()

	if err = gomigrate.Migrate(ctx, sqlDB, postgres.New, migrations,
		gomigrate.WithDirectory("migrations"),
		gomigrate.WithLogger(slog.Default()),
	); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}
	return nil
}

type DB struct {
	pool *pgxpool.Pool
	*dbsqlc.Queries
}

func (d *DB) Close() error {
	d.pool.Close()
	return nil
}
