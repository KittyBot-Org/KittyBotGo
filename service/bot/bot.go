package bot

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"sync"
	"time"

	"net/http"

	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/disgolink/v4/disgolink"
	"github.com/disgoorg/lavaqueue-plugin"
	"github.com/disgoorg/snowflake/v2"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/topi314/tint"

	"github.com/kittybot-org/kittybotgo/internal/cacheclient"
	"github.com/kittybot-org/kittybotgo/internal/otel"
	"github.com/kittybot-org/kittybotgo/service/bot/db"
	"github.com/kittybot-org/kittybotgo/service/bot/db/dbsqlc"
)

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

	if cfg.Bot.CacheURL == "" {
		cancel()
		return nil, fmt.Errorf("cache url is required")
	}
	remote, err := newRemoteGateway(cfg.Bot.NATSURL)
	if err != nil {
		cancel()
		return nil, err
	}

	var restClientConfigOpts []rest.ClientConfigOpt
	if cfg.Bot.RestURL != "" {
		restClientConfigOpts = append(restClientConfigOpts,
			rest.WithURL(cfg.Bot.RestURL),
			rest.WithRateLimiter(rest.NewNoopRateLimiter()),
		)
	}
	restClientConfigOpts = append(restClientConfigOpts, rest.WithHTTPClient(&http.Client{
		Transport: otel.Transport(nil),
	}))

	d, err := disgo.New(cfg.Bot.Token,
		bot.WithGateway(remote),
		bot.WithCaches(cacheclient.New(cfg.Bot.CacheURL)),
		bot.WithRestClientConfigOpts(restClientConfigOpts...),
		bot.WithEventListenerFunc(b.OnDiscordEvent),
	)
	if err != nil {
		cancel()
		remote.Close(context.Background())
		return nil, fmt.Errorf("failed to create discord client: %w", err)
	}
	remote.bind(d)

	lavalink := disgolink.New(d.ApplicationID,
		disgolink.WithPlugins(lavaqueue.New()),
		disgolink.WithListenerFunc(b.OnLavalinkEvent),
	)

	database, err := db.New(cfg.Database)
	if err != nil {
		cancel()
		remote.Close(context.Background())
		return nil, fmt.Errorf("failed to create database: %w", err)
	}

	b.Discord = d
	b.remote = remote
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
	remote     *remoteGateway
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

	return b.Discord.OpenGateway(context.Background())
}

func (b *Bot) ConnectLavalinkNodes() {
	nodes, err := b.Database.ListLavalinkNodes(context.Background())
	if err != nil {
		slog.Error("failed to get lavalink node session ids", tint.Err(err))
	}

	var wg sync.WaitGroup
	for i := range b.Config.Nodes {
		wg.Add(1)
		cfg := b.Config.Nodes[i]
		go func() {
			defer wg.Done()

			nodeIndex := slices.IndexFunc(nodes, func(node dbsqlc.ListLavalinkNodesRow) bool {
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
		for node := range b.Lavalink.Nodes() {
			if err := b.Database.UpsertLavalinkNode(context.Background(), dbsqlc.UpsertLavalinkNodeParams{
				Name:      node.Config.Name,
				SessionID: pgtype.Text{String: node.Config.SessionID, Valid: true},
			}); err != nil {
				slog.Error("failed to set lavalink node session ids", tint.Err(err))
			}
		}
		b.Lavalink.Close()
	}

	b.Discord.Close(context.Background())
	_ = b.Database.Close()
}
