package types

import (
	"github.com/DisgoOrg/disgolink/lavalink"
	"github.com/KittyBot-Org/KittyBotGo/internal/shared"
)

type Config struct {
	shared.Config
	Bot                  BotConfig      `json:"bot"`
	PlayHistoryCacheSize int            `json:"play_history_cache_size"`
	Lavalink             LavalinkConfig `json:"lavalink"`
	MetricsAddress       string         `json:"metrics_address"`
}

type BotConfig struct {
	ShardIDs   []int `json:"shard_ids"`
	ShardCount int   `json:"shard_count"`
}

type LavalinkConfig struct {
	Nodes           []lavalink.NodeConfig `json:"nodes"`
	ResumingTimeout int                   `json:"resuming_timeout"`
}
