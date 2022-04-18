package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/KittyBot-Org/KittyBotGo/internal/config"
	"github.com/KittyBot-Org/KittyBotGo/internal/db"
	"github.com/KittyBot-Org/KittyBotGo/internal/i18n"
	"github.com/KittyBot-Org/KittyBotGo/internal/kbot"
	"github.com/KittyBot-Org/KittyBotGo/internal/metrics"
	"github.com/KittyBot-Org/KittyBotGo/internal/modules"
	"github.com/disgoorg/log"
	_ "github.com/lib/pq"
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
	logger.Info("Starting discord kbot version: ", version)

	var cfg kbot.Config
	if err := config.LoadConfig(&cfg); err != nil {
		logger.Fatal("Failed to load config: ", err)
	}
	logger.SetLevel(cfg.LogLevel)

	logger.Info("Syncing commands? ", *shouldSyncCommands)
	logger.Info("Syncing DB tables? ", *shouldSyncDBTables)
	logger.Info("Exiting after syncing? ", *exitAfterSync)
	defer logger.Info("Shutting down discord kbot...")

	b := &kbot.Bot{
		Logger:  logger,
		Config:  cfg,
		Version: version,
	}

	if err := i18n.Setup(b); err != nil {
		b.Logger.Fatal("Failed to setup i18n: ", err)
	}

	b.LoadModules(modules.Modules)
	b.SetupPaginator()

	if err := b.SetupBot(); err != nil {
		b.Logger.Fatal("Failed to setup discord kbot: ", err)
	}
	defer b.Client.Close(context.TODO())

	if *shouldSyncCommands {
		b.SyncCommands()
	}

	var err error
	if b.DB, err = db.SetupDatabase(b.Config.Database); err != nil {
		b.Logger.Fatal("Failed to setup database: ", err)
	}
	defer b.DB.Close()

	if *exitAfterSync {
		b.Logger.Infof("Exiting after syncing commands and database tables")
		os.Exit(0)
	}

	metrics.Setup(b)

	b.SetupLavalink()
	defer b.Lavalink.Close()

	if err = b.StartBot(); err != nil {
		b.Logger.Fatal("Failed to start discord kbot: ", err)
	}

	b.Logger.Info("Bot is running. Press CTRL-C to exit.")
	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-s
}
