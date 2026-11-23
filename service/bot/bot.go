package bot

import (
	"context"
	_ "embed"
	"fmt"
	"log/slog"
	"slices"
	"sync"
	"time"

	"net/http"

	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/disgo/sharding"
	"github.com/disgoorg/disgolink/v4/disgolink"
	"github.com/disgoorg/lavaqueue-plugin"
	"github.com/disgoorg/snowflake/v2"
	"github.com/topi314/tint"

	"github.com/kittybot-org/kittybotgo/service/bot/db"
)

//go:embed sql/schema.sql
var schema string

func New(cfg Config, version string, commit string) (*Bot, error) {
	ctx, cancel := context.WithCancel(context.Background())
	b := &Bot{
		Config:  cfg,
		Version: version,
		Commit:  commit,
		ctx:     ctx,
		cancel:  cancel,
		invites: map[snowflake.ID]map[string]int{},
		used:    map[snowflake.ID]map[snowflake.ID]string{},
		states:  map[string]time.Time{},
	}

	gatewayConfigOpts := []gateway.ConfigOpt{
		gateway.WithIntents(
			gateway.IntentGuilds,
			gateway.IntentGuildVoiceStates,
			gateway.IntentGuildMembers,
			gateway.IntentGuildInvites,
		),
	}
	shardManagerConfigOpts := []sharding.ConfigOpt{
		sharding.WithGatewayConfigOpts(gatewayConfigOpts...),
	}
	if cfg.Bot.GatewayURL != "" {
		shardManagerConfigOpts = []sharding.ConfigOpt{
			sharding.WithGatewayConfigOpts(append(gatewayConfigOpts,
				gateway.WithURL(cfg.Bot.GatewayURL),
				gateway.WithCompression(gateway.CompressionNone),
			)...),
		}
	}

	var restClientConfigOpts []rest.ClientConfigOpt
	if cfg.Bot.RestURL != "" {
		restClientConfigOpts = []rest.ClientConfigOpt{
			rest.WithURL(cfg.Bot.RestURL),
			rest.WithRateLimiter(rest.NewNoopRateLimiter()),
		}
	}

	d, err := disgo.New(cfg.Bot.Token,
		bot.WithShardManagerConfigOpts(shardManagerConfigOpts...),
		bot.WithRestClientConfigOpts(restClientConfigOpts...),
		bot.WithCacheConfigOpts(
			cache.WithCaches(cache.FlagGuilds, cache.FlagVoiceStates, cache.FlagChannels),
		),
		bot.WithEventListenerFunc(b.OnDiscordEvent),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create discord client: %w", err)
	}

	lavalink := disgolink.New(d.ApplicationID,
		disgolink.WithPlugins(lavaqueue.New()),
		disgolink.WithListenerFunc(b.OnLavalinkEvent),
	)

	database, err := db.New(cfg.Database, schema)
	if err != nil {
		return nil, fmt.Errorf("failed to create database: %w", err)
	}

	b.Discord = d
	b.Database = database
	b.Lavalink = lavalink
	return b, nil
}

type Bot struct {
	Config     Config
	Version    string
	Commit     string
	Started    time.Time
	Discord    *bot.Client
	Lavalink   *disgolink.Client
	Database   *db.DB
	ctx        context.Context
	cancel     context.CancelFunc
	httpServer *http.Server

	inviteMu sync.Mutex
	invites  map[snowflake.ID]map[string]int
	usedMu   sync.Mutex
	used     map[snowflake.ID]map[snowflake.ID]string

	stateMu sync.Mutex
	states  map[string]time.Time
}

func (b *Bot) Start(commands []discord.ApplicationCommandCreate) error {
	if b.Config.Bot.SyncCommands {
		if err := handler.SyncCommands(b.Discord, commands, b.Config.Bot.GuildIDs); err != nil {
			slog.Error("failed to sync commands", tint.Err(err))
		}
	}

	b.Started = time.Now()
	b.ConnectLavalinkNodes()
	go b.scheduleNotifications()
	b.startDashboard()

	return b.Discord.OpenShardManager(context.Background())
}

func (b *Bot) ConnectLavalinkNodes() {
	nodes, err := b.Database.GetLavalinkNodes(context.Background())
	if err != nil {
		slog.Error("failed to get lavalink node session ids", tint.Err(err))
	}

	var wg sync.WaitGroup
	for i := range b.Config.Nodes {
		wg.Add(1)
		cfg := b.Config.Nodes[i]
		go func() {
			defer wg.Done()

			nodeIndex := slices.IndexFunc(nodes, func(node db.LavalinkNode) bool {
				return node.Name == cfg.Name
			})

			var sessionID string
			if nodeIndex > -1 {
				sessionID = nodes[nodeIndex].SessionID
			}

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			if _, err = b.Lavalink.AddNode(ctx, cfg.ToLavalink(sessionID)); err != nil {
				slog.Error("failed to add node", tint.Err(err))
			}
			slog.Info("connected to lavalink node", slog.String("name", cfg.Name))
		}()
	}
	wg.Wait()
}

func (b *Bot) Close() {
	if b.cancel != nil {
		b.cancel()
	}
	if b.httpServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = b.httpServer.Shutdown(ctx)
	}
	if b.Lavalink != nil {
		var nodes []db.LavalinkNode
		for node := range b.Lavalink.Nodes() {
			nodes = append(nodes, db.LavalinkNode{
				Name:      node.Config.Name,
				SessionID: node.Config.SessionID,
			})
		}

		if len(nodes) > 0 {
			if err := b.Database.AddLavalinkNodes(context.Background(), nodes); err != nil {
				slog.Error("failed to set lavalink node session ids", tint.Err(err))
			}
		}
		b.Lavalink.Close()
	}

	b.Discord.Close(context.Background())
	_ = b.Database.Close()
}
