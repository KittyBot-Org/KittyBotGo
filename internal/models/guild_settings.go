package models

import "github.com/DisgoOrg/snowflake"

type GuildSettings struct {
	ID snowflake.Snowflake `bun:"id,pk,notnull"`
}
