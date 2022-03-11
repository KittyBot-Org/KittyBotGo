package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/DisgoOrg/log"
	"github.com/KittyBot-Org/KittyBotGo/internal/bot/i18n"
	"github.com/KittyBot-Org/KittyBotGo/internal/bot/metrics"
	"github.com/KittyBot-Org/KittyBotGo/internal/bot/types"
	"github.com/KittyBot-Org/KittyBotGo/internal/shared"
	"github.com/KittyBot-Org/KittyBotGo/modules"
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
	var err error
	logger := log.New(log.Ldate | log.Ltime | log.Lshortfile)

	bot := &types.Bot{
		Logger:  logger,
		Version: version,
	}
	bot.Logger.Infof("Starting bot version: %s", version)
	bot.Logger.Infof("Syncing commands? %v", *shouldSyncCommands)
	bot.Logger.Infof("Syncing DB tables? %v", *shouldSyncDBTables)
	bot.Logger.Infof("Exiting after syncing? %v", *exitAfterSync)
	defer bot.Logger.Info("Shutting down bot...")

	if err = shared.LoadConfig(&bot.Config); err != nil {
		bot.Logger.Fatal("Failed to load config: ", err)
	}
	logger.SetLevel(bot.Config.LogLevel)

	if err = i18n.Setup(bot); err != nil {
		bot.Logger.Fatal("Failed to setup i18n: ", err)
	}

	bot.LoadModules(modules.Modules)
	bot.SetupPaginator()

	if err = bot.SetupBot(); err != nil {
		bot.Logger.Fatal("Failed to setup bot: ", err)
	}
	defer bot.Bot.Close(context.TODO())

	if *shouldSyncCommands {
		bot.SyncCommands()
	}

	if bot.DB, err = shared.SetupDatabase(bot.Config.Database, *shouldSyncDBTables, bot.Config.DevMode); err != nil {
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
	defer bot.SavePlayers()

	if err = bot.StartBot(); err != nil {
		bot.Logger.Fatal("Failed to start bot: ", err)
	}

	bot.Logger.Info("Bot is running. Press CTRL-C to exit.")
	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-s
}
