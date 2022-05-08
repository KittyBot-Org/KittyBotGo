package db

import (
	"database/sql"

	. "github.com/KittyBot-Org/KittyBotGo/internal/db/.gen/kittybot-go/public/model"
	"github.com/disgoorg/snowflake/v2"
)

type GuildSettingsDB interface {
	Get(guildID snowflake.ID) (GuildSetting, error)
	Set(model GuildSetting) error
	Delete(guildID snowflake.ID) error
}

type guildSettingsDBImpl struct {
	db *sql.DB
}

func (s *guildSettingsDBImpl) Get(guildID snowflake.ID) (GuildSetting, error) {
	return GuildSetting{}, nil
}

func (s *guildSettingsDBImpl) Set(model GuildSetting) error {
	return nil
}

func (s *guildSettingsDBImpl) Delete(guildID snowflake.ID) error {
	return nil
}
