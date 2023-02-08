package bot

import (
	"github.com/KittyBot-Org/KittyBotGo/config"
	"github.com/KittyBot-Org/KittyBotGo/database"
)

type Config struct {
	DevMode  bool            `json:"dev_mode"`
	Database database.Config `json:"database"`
	Nats     config.NATS     `json:"nats"`
}
