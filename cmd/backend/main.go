package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	backend2 "github.com/KittyBot-Org/KittyBotGo/backend"
	"github.com/KittyBot-Org/KittyBotGo/backend/routes"
	"github.com/KittyBot-Org/KittyBotGo/config"
	"github.com/KittyBot-Org/KittyBotGo/db"
	"github.com/KittyBot-Org/KittyBotGo/dbot/commands"
	"github.com/disgoorg/log"
	_ "github.com/lib/pq"
)

var (
	shouldSyncDBTables *bool
	exitAfterSync      *bool
	version            = "dev"
)

func init() {
	shouldSyncDBTables = flag.Bool("sync-db", false, "Whether to sync the database tables")
	exitAfterSync = flag.Bool("exit-after-sync", false, "Whether to exit after syncing commands and database tables")
	flag.Parse()
}

func main() {
	var err error
	logger := log.New(log.Ldate | log.Ltime | log.Lshortfile)

	logger.Infof("Starting b version: %s", version)
	logger.Infof("Syncing DB tables? %v", *shouldSyncDBTables)
	logger.Infof("Exiting after syncing? %v", *exitAfterSync)

	var cfg backend2.Config
	if err = config.LoadConfig(&cfg); err != nil {
		logger.Fatal("Failed to load config: ", err)
	}
	logger.SetLevel(cfg.LogLevel)

	b := &backend2.Backend{
		Logger:  logger,
		Config:  cfg,
		Version: version,
	}

	if b.DB, err = db.SetupDatabase(b.Config.Database); err != nil {
		b.Logger.Fatal("Failed to setup database: ", err)
	}
	defer b.DB.Close()

	if *exitAfterSync {
		b.Logger.Infof("Exiting after syncing database tables")
		os.Exit(0)
	}

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
		commands.Settings,
	)
	b.SetupRestServices()
	if err = b.SetupPrometheusAPI(); err != nil {
		b.Logger.Fatal("Failed to setup prometheus api: ", err)
	}
	if err = b.SetupScheduler(); err != nil {
		b.Logger.Fatal("Failed to setup scheduler: ", err)
	}
	defer b.Scheduler.Shutdown()
	b.SetupServer(routes.Handler(b))

	b.Logger.Info("Backend is running. Press CTRL-C to exit.")
	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-s
}
