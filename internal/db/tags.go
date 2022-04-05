package db

import (
	"time"

	"github.com/disgoorg/snowflake"
)

type TagsDB interface {
	Get(guildID snowflake.Snowflake, name string) (TagModel, error)
	GetAll(guildID snowflake.Snowflake) ([]TagModel, error)
	Set(model TagModel) error
	Delete(guildID snowflake.Snowflake, name string) error
}

type TagModel struct {
	GuildID   snowflake.Snowflake `bun:"guild_id,pk"`
	OwnerID   snowflake.Snowflake `bun:"owner_id,notnull"`
	Name      string              `bun:"name,pk"`
	Content   string              `bun:"content,notnull"`
	Uses      int                 `bun:"uses,default:0"`
	CreatedAt time.Time           `bun:"created_at,nullzero,notnull,default:current_timestamp"`
}
