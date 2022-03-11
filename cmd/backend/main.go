package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/DisgoOrg/log"
	"github.com/KittyBot-Org/KittyBotGo/internal/backend/routes"
	"github.com/KittyBot-Org/KittyBotGo/internal/backend/types"
	"github.com/KittyBot-Org/KittyBotGo/internal/shared"
	"github.com/KittyBot-Org/KittyBotGo/modules"
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
	backend := &types.Backend{
		Logger:  logger,
		Version: version,
	}
	backend.Logger.Infof("Starting backend version: %s", version)
	backend.Logger.Infof("Syncing DB tables? %v", *shouldSyncDBTables)
	backend.Logger.Infof("Exiting after syncing? %v", *exitAfterSync)

	if err = shared.LoadConfig(&backend.Config); err != nil {
		backend.Logger.Fatal("Failed to load config: ", err)
	}
	logger.SetLevel(backend.Config.LogLevel)

	if backend.DB, err = shared.SetupDatabase(backend.Config.Database, *shouldSyncDBTables, backend.Config.DevMode); err != nil {
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
