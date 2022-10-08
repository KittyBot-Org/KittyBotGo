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
	"github.com/disgoorg/log"
	"github.com/disgoorg/snowflake/v2"
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

	b := dbot.New(logger, cfg, version)
	b.Handler.AddCommands(
		commands.BassBoost(b),
		commands.ClearQueue(b),
		commands.History(b),
		commands.LikedSongs(b),
		commands.Loop(b),
		commands.Next(b),
		commands.NowPlaying(b),
		commands.Pause(b),
		commands.Play(b),
		commands.Previous(b),
		commands.Queue(b),
		commands.Remove(b),
		commands.Seek(b),
		commands.Shuffle(b),
		commands.Stop(b),
		commands.Tag(b),
		commands.Tags(b),
		commands.Volume(b),
		commands.Report(b),
		commands.Reports(b),
		commands.ReportUser(b),
		commands.Settings(b),
	)

	b.Handler.AddComponents(
		commands.ReportAction(b),
		commands.ReportConfirm(b),
		commands.ReportDelete(b),
		commands.PlayerLike(b),
		commands.PlayerNext(b),
		commands.PlayerPlayPause(b),
		commands.PlayerPrevious(b),
	)

	b.Handler.AddModals(
		commands.ReportActionConfirm(b),
	)

	if err := b.SetupBot(
		listeners.Metrics(b),
		listeners.Moderation(b),
		listeners.Music(b),
		listeners.Settings(b),
	); err != nil {
		b.Logger.Fatal("Failed to setup discord dbot: ", err)
	}
	defer b.Client.Close(context.TODO())

	if *shouldSyncCommands {
		var guilds []snowflake.ID
		if b.Config.DevMode {
			guilds = b.Config.DevGuildIDs
		}
		b.Handler.SyncCommands(b.Client, guilds...)
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
