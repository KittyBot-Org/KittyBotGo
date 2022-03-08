package types

import (
	"github.com/DisgoOrg/disgolink/lavalink"
	"github.com/KittyBot-Org/KittyBotGo/internal/config"
	"github.com/KittyBot-Org/KittyBotGo/internal/database"
)

type Config struct {
	config.Config
	Database             database.Config `json:"database"`
	PlayHistoryCacheSize int             `json:"play_history_cache_size"`
	Lavalink             LavalinkConfig  `json:"lavalink"`
}

type LavalinkConfig struct {
	Nodes           []lavalink.NodeConfig `json:"nodes"`
	ResumingTimeout int                   `json:"resuming_timeout"`
}
