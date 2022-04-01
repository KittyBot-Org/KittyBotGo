package db

import (
	"time"

	"github.com/disgoorg/snowflake"
)

type PlayHistory interface {
	Get(userID snowflake.Snowflake) ([]PlayHistoryModel, error)
	Add(model PlayHistoryModel) error
}

type PlayHistoryModel struct {
	UserID     snowflake.Snowflake `bun:"user_id,pk"`
	Query      string              `bun:"query,notnull"`
	Title      string              `bun:"title,pk"`
	LastUsedAt time.Time           `bun:"last_used_at,nullzero,notnull,default:current_timestamp"`
}
