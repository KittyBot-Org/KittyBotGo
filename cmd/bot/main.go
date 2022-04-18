package main

import (
	"context"
	"flag"
	"github.com/KittyBot-Org/KittyBotGo/internal/dbot"
	"github.com/KittyBot-Org/KittyBotGo/internal/i18n"
	"github.com/KittyBot-Org/KittyBotGo/internal/metrics"
	"github.com/KittyBot-Org/KittyBotGo/internal/modules"
	"os"
	"os/signal"
	"syscall"

	"github.com/KittyBot-Org/KittyBotGo/internal/config"
	"github.com/KittyBot-Org/KittyBotGo/internal/db"
	"github.com/disgoorg/log"
)

var (
	shouldSyncCommands *bool
	shouldSyncDBTables *bool
	exitAfterSync      *bool
	version            = "dev"
)

func init() {
	shouldSyncCommands = flag.Bool("sync-commands", false, "Whether to sync commands to discord")
	shouldSyncDBTables = flag.Bool("sync-db", false, "Whether to sync the database tables")
	exitAfterSync = flag.Bool("exit-after-sync", false, "Whether to exit after syncing commands and database tables")
	flag.Parse()
}

func main() {
	logger := log.New(log.Ldate | log.Ltime | log.Lshortfile)
	logger.Info("Starting discord bot version: ", version)

	var cfg dbot.Config
	if err := config.LoadConfig(&cfg); err != nil {
		logger.Fatal("Failed to load config: ", err)
	}
	logger.SetLevel(cfg.LogLevel)

	logger.Info("Syncing commands? ", *shouldSyncCommands)
	logger.Info("Syncing DB tables? ", *shouldSyncDBTables)
	logger.Info("Exiting after syncing? ", *exitAfterSync)
	defer logger.Info("Shutting down discord bot...")

	bot := &dbot.Bot{
		Logger:  logger,
		Version: version,
	}

	if err := i18n.Setup(bot); err != nil {
		bot.Logger.Fatal("Failed to setup i18n: ", err)
	}

	bot.LoadModules(modules.Modules)
	bot.SetupPaginator()

	if err := bot.SetupBot(); err != nil {
		bot.Logger.Fatal("Failed to setup discord bot: ", err)
	}
	defer bot.Client.Close(context.TODO())

	if *shouldSyncCommands {
		bot.SyncCommands()
	}

	var err error
	if bot.DB, err = db.SetupDatabase(bot.Config.Database); err != nil {
		bot.Logger.Fatal("Failed to setup database: ", err)
	}
	defer bot.DB.Close()

	if *exitAfterSync {
		bot.Logger.Infof("Exiting after syncing commands and database tables")
		os.Exit(0)
	}

	metrics.Setup(bot)

	bot.SetupLavalink()
	defer bot.Lavalink.Close()

	if err = bot.StartBot(); err != nil {
		bot.Logger.Fatal("Failed to start discord bot: ", err)
	}

	bot.Logger.Info("Bot is running. Press CTRL-C to exit.")
	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-s
}
