package main

import (
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/topi314/tint"

	"github.com/KittyBot-Org/KittyBotGo/internal/config"
	"github.com/KittyBot-Org/KittyBotGo/internal/log"
	"github.com/KittyBot-Org/KittyBotGo/service/bot"
	"github.com/KittyBot-Org/KittyBotGo/service/bot/commands"
)

var (
	Version = "dev"
	Commit  = "unknown"
)

func main() {
	cfgPath := flag.String("config", "config.toml", "path to config file")
	flag.Parse()

	slog.Info("Bot is starting...", slog.String("config", *cfgPath))

	var cfg bot.Config
	if err := config.Load(*cfgPath, &cfg); err != nil {
		slog.Error("failed to load config", tint.Err(err))
		return
	}

	slog.Info("Config loaded", slog.String("config", cfg.String()))
	log.Setup(cfg.Log)

	b, err := bot.New(cfg, Version, Commit)
	if err != nil {
		slog.Error("Failed to create bot: %v", err)
	}
	defer b.Close()

	b.Discord.AddEventListeners(commands.New(b))

	if err = b.Start(commands.Commands); err != nil {
		slog.Error("Failed to start bot: %v", err)
		return
	}

	slog.Info("Bot is running. Press CTRL-C to exit.")
	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM)
	<-s
}
