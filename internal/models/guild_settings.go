package models

import "github.com/disgoorg/snowflake"

type GuildSettings struct {
	ID snowflake.Snowflake `bun:"id,pk,notnull"`
}
