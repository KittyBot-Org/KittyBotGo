package models

import (
	"time"

	"github.com/DisgoOrg/snowflake"
)

type PlayHistory struct {
	UserID     snowflake.Snowflake `bun:"user_id,pk"`
	Query      string              `bun:"query,notnull"`
	Title      string              `bun:"title,pk"`
	LastUsedAt time.Time           `bun:"last_used_at,nullzero,notnull,default:current_timestamp"`
}
