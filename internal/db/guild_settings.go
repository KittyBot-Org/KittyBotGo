package db

import "github.com/disgoorg/snowflake"

type GuildSettings interface {
	Get(guildID snowflake.Snowflake) (GuildSettingsModel, error)
	Set(model GuildSettingsModel) error
	Delete(guildID snowflake.Snowflake) error
}

type GuildSettingsModel struct {
	ID snowflake.Snowflake `bun:"id,pk,notnull"`
}
