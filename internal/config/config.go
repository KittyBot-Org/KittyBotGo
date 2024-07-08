package config

import (
	"fmt"
	"os"

	"github.com/pelletier/go-toml/v2"
)

func Load(path string, cfg any) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	return toml.NewDecoder(file).Decode(cfg)
}
