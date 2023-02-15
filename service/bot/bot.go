package bot

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/disgo/sharding"
	"github.com/disgoorg/disgolink/v2/disgolink"
	"github.com/disgoorg/disgolink/v2/lavalink"
	"github.com/disgoorg/json"
	"github.com/disgoorg/log"

	"github.com/KittyBot-Org/KittyBotGo/interal/config"
	"github.com/KittyBot-Org/KittyBotGo/interal/database"
)

func New(logger log.Logger, cfgPath string, cfg Config) (*Bot, error) {
	b := &Bot{
		CfgPath: cfgPath,
		Config:  cfg,
		Logger:  logger,
	}

	dc, err := disgo.New(cfg.Token,
		bot.WithLogger(logger),
		bot.WithShardManagerConfigOpts(
			sharding.WithGatewayConfigOpts(
				gateway.WithURL(cfg.GatewayURL),
				gateway.WithCompress(false),
			),
			sharding.WithRateLimiter(sharding.NewNoopRateLimiter()),
		),
		//bot.WithRestClientConfigOpts(
		//	rest.WithURL(cfg.RestURL),
		//	rest.WithRateLimiter(rest.NewNoopRateLimiter()),
		//),
		bot.WithCacheConfigOpts(
			cache.WithCaches(cache.FlagGuilds, cache.FlagMembers, cache.FlagVoiceStates),
		),
		bot.WithEventListenerFunc(b.OnDiscordEvent),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create discord client: %w", err)
	}

	ll := disgolink.New(dc.ApplicationID(),
		disgolink.WithLogger(logger),
		disgolink.WithListenerFunc(b.OnLavalinkEvent),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	db, err := database.New(ctx, cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to create database: %w", err)
	}

	b.Discord = dc
	b.Database = db
	b.Lavalink = ll
	return b, nil
}

type Bot struct {
	CfgPath  string
	Config   Config
	Logger   log.Logger
	Discord  bot.Client
	Lavalink disgolink.Client
	Database *database.Database
}

func (b *Bot) Start(commands []discord.ApplicationCommandCreate) error {
	if b.Config.SyncCommands {
		if b.Config.DevMode {
			b.Logger.Info("starting in dev mode")
			for _, guildID := range b.Config.GuildIDs {
				if _, err := b.Discord.Rest().SetGuildCommands(b.Discord.ApplicationID(), guildID, commands); err != nil {
					return fmt.Errorf("failed to update guild handlers: %w", err)
				}
			}
		} else {
			if _, err := b.Discord.Rest().SetGlobalCommands(b.Discord.ApplicationID(), commands); err != nil {
				return fmt.Errorf("failed to update global handlers: %w", err)
			}
		}
	}

	var wg sync.WaitGroup
	for i := range b.Config.Nodes {
		wg.Add(1)
		cfg := b.Config.Nodes[i]
		go func() {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			node, err := b.Lavalink.AddNode(ctx, cfg)
			if err != nil {
				b.Logger.Error("failed to add node:", err)
				return
			}

			if err = node.Update(context.Background(), lavalink.SessionUpdate{
				Resuming: json.Ptr(true),
				Timeout:  json.Ptr(180),
			}); err != nil {
				b.Logger.Error("failed to update node:", err)
			}
		}()
	}
	wg.Wait()

	return b.Discord.OpenShardManager(context.Background())
}

func (b *Bot) OnDiscordEvent(event bot.Event) {
	switch e := event.(type) {
	case *events.VoiceServerUpdate:
		b.Logger.Debug("received voice server update")
		if e.Endpoint == nil {
			return
		}
		b.Lavalink.OnVoiceServerUpdate(context.Background(), e.GuildID, e.Token, *e.Endpoint)
	case *events.GuildVoiceStateUpdate:
		if e.VoiceState.UserID != b.Discord.ApplicationID() {
			return
		}
		b.Logger.Debug("received voice state update")
		b.Lavalink.OnVoiceStateUpdate(context.Background(), e.VoiceState.GuildID, e.VoiceState.ChannelID, e.VoiceState.SessionID)
	case *events.GuildsReady:
		b.Logger.Debug("received guilds ready")
		b.RestorePlayers()
	}
}

func (b *Bot) Close() {
	b.Lavalink.ForNodes(func(node disgolink.Node) {
		for i, cfgNode := range b.Config.Nodes {
			if node.Config().Name == cfgNode.Name {
				b.Config.Nodes[i].SessionID = node.SessionID()
			}
		}
	})

	if err := config.Save(b.CfgPath, b.Config); err != nil {
		b.Logger.Error("failed to save config:", err)
	}
	b.Lavalink.Close()
	b.Discord.Close(context.Background())
	_ = b.Database.Close()
}
