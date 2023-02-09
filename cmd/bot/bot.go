package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/log"

	"github.com/KittyBot-Org/KittyBotGo/interal/config"
	"github.com/KittyBot-Org/KittyBotGo/service/bot"
	"github.com/KittyBot-Org/KittyBotGo/service/bot/commands"
)

func main() {
	cfgPath := flag.String("config", "config.json", "path to config.json")
	flag.Parse()

	logger := log.New(log.Ldate | log.Ltime | log.Lshortfile)
	logger.Infof("Bot is starting... (config path:%s)", *cfgPath)

	var cfg bot.Config
	if err := config.Load(*cfgPath, &cfg); err != nil {
		logger.Fatalf("Failed to load config: %v", err)
	}
	logger.SetLevel(config.ParseLogLevel(cfg.LogLevel))

	b, err := bot.New(logger, cfg)
	if err != nil {
		logger.Fatalf("Failed to create bot: %v", err)
	}

	cmds := commands.New(b)
	r := handler.New()
	r.HandleCommand("/ping", cmds.OnPing)
	r.HandleCommand("/play", cmds.OnPlay)
	b.Discord.AddEventListeners(r)

	if err = b.Start(commands.Commands); err != nil {
		logger.Fatalf("Failed to start bot: %v", err)
	}

	logger.Info("Bot is running. Press CTRL-C to exit.")
	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-s
}
