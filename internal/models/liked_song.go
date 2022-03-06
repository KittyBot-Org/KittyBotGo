package models

import (
	"github.com/DisgoOrg/snowflake"
	"time"
)

type LikedSong struct {
	ID        int                 `bun:"id,autoincrement,pk,notnull"`
	UserID    snowflake.Snowflake `bun:"user_id,notnull,unique:user-song"`
	Query     string              `bun:"query,notnull"`
	Title     string              `bun:"title,notnull,unique:user-song"`
	CreatedAt time.Time           `bun:"created_at,nullzero,notnull,default:current_timestamp"`
}
