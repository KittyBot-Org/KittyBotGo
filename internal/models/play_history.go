package models

import (
	"time"

	"github.com/DisgoOrg/snowflake"
)

type PlayHistory struct {
	ID         int                 `bun:"id,autoincrement,pk,notnull"`
	UserID     snowflake.Snowflake `bun:"user_id,notnull"`
	Query      string              `bun:"query,notnull"`
	Title      string              `bun:"title,notnull"`
	LastUsedAt time.Time           `bun:"last_used_at,nullzero,notnull,default:current_timestamp"`
}
