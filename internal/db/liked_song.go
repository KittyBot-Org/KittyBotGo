package db

import (
	"time"

	"github.com/disgoorg/snowflake"
)

type LikedSongs interface {
	Get(userID snowflake.Snowflake) ([]LikedSongModel, error)
	Add(model LikedSongModel) error
	Delete(userID snowflake.Snowflake, title string) error
}

type LikedSongModel struct {
	UserID    snowflake.Snowflake `bun:"user_id,pk"`
	Query     string              `bun:"query,notnull"`
	Title     string              `bun:"title,pk"`
	CreatedAt time.Time           `bun:"created_at,nullzero,notnull,default:current_timestamp"`
}
