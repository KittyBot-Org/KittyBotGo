package models

import (
	"time"

	"github.com/DisgoOrg/snowflake"
)

type LikedSong struct {
	UserID    snowflake.Snowflake `bun:"user_id,pk"`
	Query     string              `bun:"query,notnull"`
	Title     string              `bun:"title,pk"`
	CreatedAt time.Time           `bun:"created_at,nullzero,notnull,default:current_timestamp"`
}
