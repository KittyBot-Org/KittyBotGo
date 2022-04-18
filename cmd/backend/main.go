package main

import (
	"flag"
	"github.com/KittyBot-Org/KittyBotGo/internal/bend"
	"github.com/KittyBot-Org/KittyBotGo/internal/modules"
	"github.com/KittyBot-Org/KittyBotGo/internal/routes"
	"os"
	"os/signal"
	"syscall"

	"github.com/KittyBot-Org/KittyBotGo/internal/config"
	"github.com/KittyBot-Org/KittyBotGo/internal/db"
	"github.com/disgoorg/log"
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

	logger.Infof("Starting bend version: %s", version)
	logger.Infof("Syncing DB tables? %v", *shouldSyncDBTables)
	logger.Infof("Exiting after syncing? %v", *exitAfterSync)

	var cfg bend.Config
	if err = config.LoadConfig(&cfg); err != nil {
		logger.Fatal("Failed to load config: ", err)
	}
	logger.SetLevel(cfg.LogLevel)

	backend := &bend.Backend{
		Logger:  logger,
		Version: version,
	}

	if backend.DB, err = db.SetupDatabase(backend.Config.Database); err != nil {
		backend.Logger.Fatal("Failed to setup database: ", err)
	}
	defer backend.DB.Close()

	if *exitAfterSync {
		backend.Logger.Infof("Exiting after syncing database tables")
		os.Exit(0)
	}

	backend.LoadCommands(modules.Modules)
	backend.SetupRestServices()
	if err = backend.SetupPrometheusAPI(); err != nil {
		backend.Logger.Fatal("Failed to setup prometheus api: ", err)
	}
	if err = backend.SetupScheduler(); err != nil {
		backend.Logger.Fatal("Failed to setup scheduler: ", err)
	}
	defer backend.Scheduler.Shutdown()
	backend.SetupServer(routes.Handler(backend))

	backend.Logger.Info("Backend is running. Press CTRL-C to exit.")
	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-s
}
