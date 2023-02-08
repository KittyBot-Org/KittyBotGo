package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/disgoorg/json"
	"github.com/disgoorg/log"
)

type NATS struct {
	URL      string `json:"url"`
	User     string `json:"user"`
	Password string `json:"password"`
	Queue    string `json:"queue"`
}

func (n NATS) String() string {
	return fmt.Sprintf("\n  URL: %s,\n  User: %s,\n  Password: %s,\n Queue: %s", n.URL, n.User, strings.Repeat("*", len(n.Password)), n.Queue)
}

func ParseLogLevel(level string) log.Level {
	logLevel := log.LevelInfo
	switch strings.ToLower(level) {
	case "debug":
		logLevel = log.LevelDebug
	case "info":
		logLevel = log.LevelInfo
	case "warn":
		logLevel = log.LevelWarn
	case "error":
		logLevel = log.LevelError
	case "fatal":
		logLevel = log.LevelFatal
	case "panic":
		logLevel = log.LevelPanic
	}

	return logLevel
}

func Load(path string, cfg any) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	return json.NewDecoder(file).Decode(cfg)
}
