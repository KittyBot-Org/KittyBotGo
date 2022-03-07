package types

import (
	"github.com/KittyBot-Org/KittyBotGo/internal/config"
	"github.com/KittyBot-Org/KittyBotGo/internal/database"
)

type Config struct {
	config.Config
	Backend  BackendConfig   `json:"backend"`
	Database database.Config `json:"database"`

	BotInvite   string `json:"bot_invite"`
	GuildInvite string `json:"guild_invite"`
}

type BackendConfig struct {
	Token string `json:"token"`
	Port  string `json:"port"`
}
