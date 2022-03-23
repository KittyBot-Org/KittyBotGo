package models

import (
	"time"

	"github.com/disgoorg/snowflake"
)

type Tag struct {
	ID        int                 `bun:"id,autoincrement,pk,notnull"`
	GuildID   snowflake.Snowflake `bun:"guild_id,notnull,unique:name-guild"`
	OwnerID   snowflake.Snowflake `bun:"owner_id,notnull"`
	Name      string              `bun:"name,notnull,unique:name-guild"`
	Content   string              `bun:"content,notnull"`
	Uses      int                 `bun:"uses,default:0"`
	CreatedAt time.Time           `bun:"created_at,nullzero,notnull,default:current_timestamp"`
}
