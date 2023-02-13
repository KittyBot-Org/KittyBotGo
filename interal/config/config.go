package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/disgoorg/json"
	"github.com/disgoorg/log"
)

func ParseLogLevel(level string) log.Level {
	logLevel := log.LevelInfo
	switch strings.ToLower(level) {
	case "trace":
		logLevel = log.LevelTrace
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

func Save(path string, cfg any) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "\t")
	return encoder.Encode(cfg)
}
