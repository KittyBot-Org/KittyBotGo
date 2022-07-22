package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/KittyBot-Org/KittyBotGo/config"
	"github.com/KittyBot-Org/KittyBotGo/db"
	"github.com/KittyBot-Org/KittyBotGo/dbot"
	"github.com/KittyBot-Org/KittyBotGo/dbot/commands"
	"github.com/KittyBot-Org/KittyBotGo/dbot/listeners"
	"github.com/KittyBot-Org/KittyBotGo/i18n"
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
	logger.Info("Starting discord dbot version: ", version)

	var cfg dbot.Config
	if err := config.LoadConfig(&cfg); err != nil {
		logger.Fatal("Failed to load config: ", err)
	}
	logger.SetLevel(cfg.LogLevel)

	logger.Info("Syncing commands? ", *shouldSyncCommands)
	logger.Info("Syncing DB tables? ", *shouldSyncDBTables)
	logger.Info("Exiting after syncing? ", *exitAfterSync)
	defer logger.Info("Shutting down discord dbot...")

	if err := i18n.Setup(logger); err != nil {
		logger.Fatal("Failed to setup i18n: ", err)
	}

	b := dbot.New(logger, cfg, version)
	b.LoadCommands(
		commands.BassBoost,
		commands.ClearQueue,
		commands.History,
		commands.LikedSongs,
		commands.Loop,
		commands.Next,
		commands.NowPlaying,
		commands.Pause,
		commands.Play,
		commands.Previous,
		commands.Queue,
		commands.Remove,
		commands.Seek,
		commands.Shuffle,
		commands.Stop,
		commands.Tag,
		commands.Tags,
		commands.Volume,
		commands.Report,
		commands.Reports,
		commands.ReportUser,
		commands.Settings,
	)

	if err := b.SetupBot(
		listeners.Metrics(b),
		listeners.Moderation(b),
		listeners.Music(b),
	); err != nil {
		b.Logger.Fatal("Failed to setup discord dbot: ", err)
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

	//metrics.Setup(b.Logger, b.Config)

	b.SetupLavalink()
	defer b.Lavalink.Close()

	if err = b.StartBot(); err != nil {
		b.Logger.Fatal("Failed to start discord dbot: ", err)
	}

	b.Logger.Info("Bot is running. Press CTRL-C to exit.")
	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-s
}
