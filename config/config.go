package config

import (
	"fmt"
	"os"

	"github.com/disgoorg/json"
)

type NATS struct {
	URL      string `json:"url"`
	User     string `json:"user"`
	Password string `json:"password"`
}

func Load(path string, cfg any) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	return json.NewDecoder(file).Decode(cfg)
}
