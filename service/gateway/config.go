package gateway

import (
	"fmt"
	"strings"

	"github.com/KittyBot-Org/KittyBotGo/interal/config"
)

type Config struct {
	Token      string      `json:"token"`
	ShardCount int         `json:"shard_count"`
	LogLevel   string      `json:"log_level"`
	Nats       config.NATS `json:"nats"`
}

func (c Config) String() string {
	return fmt.Sprintf("\n Token: %s,\n Shard Count: %d,\n Log Level: %s,\n Nats: %s\n", strings.Repeat("*", len(c.Token)), c.ShardCount, c.LogLevel, c.Nats)
}
