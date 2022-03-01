package models

import "github.com/DisgoOrg/snowflake"

type GuildSettings struct {
	GuildID snowflake.Snowflake `bun:"guild_id,pk,notnull"`
}
