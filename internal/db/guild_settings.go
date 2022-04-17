package db

import (
	"database/sql"

	. "github.com/KittyBot-Org/KittyBotGo/internal/db/.gen/kittybot-go/public/model"
	"github.com/disgoorg/snowflake"
)

type GuildSettingsDB interface {
	Get(guildID snowflake.Snowflake) (GuildSettings, error)
	Set(model GuildSettings) error
	Delete(guildID snowflake.Snowflake) error
}

type guildSettingsDBImpl struct {
	db *sql.DB
}

func (s *guildSettingsDBImpl) Get(guildID snowflake.Snowflake) (GuildSettings, error) {
	return GuildSettings{}, nil
}

func (s *guildSettingsDBImpl) Set(model GuildSettings) error {
	return nil
}

func (s *guildSettingsDBImpl) Delete(guildID snowflake.Snowflake) error {
	return nil
}
