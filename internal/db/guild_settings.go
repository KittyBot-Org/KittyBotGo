package db

import (
	. "github.com/KittyBot-Org/KittyBotGo/internal/db/.gen/kittybot-go/public/model"
	"github.com/disgoorg/snowflake"
)

type GuildSettingsDB interface {
	Get(guildID snowflake.Snowflake) (GuildSettings, error)
	Set(model GuildSettings) error
	Delete(guildID snowflake.Snowflake) error
}
