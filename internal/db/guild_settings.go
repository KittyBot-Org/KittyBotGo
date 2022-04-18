package db

import (
	"database/sql"

	. "github.com/KittyBot-Org/KittyBotGo/internal/db/.gen/kittybot-go/public/model"
	"github.com/disgoorg/snowflake"
)

type GuildSettingsDB interface {
	Get(guildID snowflake.Snowflake) (GuildSetting, error)
	Set(model GuildSetting) error
	Delete(guildID snowflake.Snowflake) error
}

type guildSettingsDBImpl struct {
	db *sql.DB
}

func (s *guildSettingsDBImpl) Get(guildID snowflake.Snowflake) (GuildSetting, error) {
	return GuildSetting{}, nil
}

func (s *guildSettingsDBImpl) Set(model GuildSetting) error {
	return nil
}

func (s *guildSettingsDBImpl) Delete(guildID snowflake.Snowflake) error {
	return nil
}
