package gateway

import (
	"github.com/KittyBot-Org/KittyBotGo/config"
)

type Config struct {
	Token string      `json:"token"`
	Nats  config.NATS `json:"nats"`
}
