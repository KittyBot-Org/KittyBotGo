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
	"github.com/disgoorg/disgolink/v2/disgolink"
	"github.com/disgoorg/log"
	"github.com/nats-io/nats.go"

	"github.com/KittyBot-Org/KittyBotGo/interal/database"
)

func New(logger log.Logger, cfg Config) (*Bot, error) {
	b := &Bot{
		Config: cfg,
		Logger: logger,
	}

	discord, err := disgo.New(cfg.Token,
		bot.WithLogger(logger),
		bot.WithGateway(b.NewNATSGateway()),
		bot.WithCacheConfigOpts(
			cache.WithCaches(cache.FlagGuilds, cache.FlagMembers, cache.FlagVoiceStates),
		),
		bot.WithEventListeners(b),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create discord client: %w", err)
	}

	lavalink := disgolink.New(discord.ApplicationID(), disgolink.WithLogger(logger))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	db, err := database.New(ctx, cfg.Database)

	conn, err := nats.Connect(cfg.Nats.URL,
		nats.Name("gateway"),
		nats.UserInfo(cfg.Nats.User, cfg.Nats.Password),
		nats.MaxReconnects(-1),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to nats: %w", err)
	}

	b.Discord = discord
	b.Database = db
	b.Lavalink = lavalink
	b.Nats = conn
	return b, nil
}

type Bot struct {
	Config   Config
	Logger   log.Logger
	Discord  bot.Client
	Lavalink disgolink.Client
	Database *database.Database
	Nats     *nats.Conn
}

func (b *Bot) Start(commands []discord.ApplicationCommandCreate) error {
	if b.Config.DevMode {
		b.Logger.Info("starting in dev mode")
		for _, guildID := range b.Config.GuildIDs {
			if _, err := b.Discord.Rest().SetGuildCommands(b.Discord.ApplicationID(), guildID, commands); err != nil {
				return fmt.Errorf("failed to update guild commands: %w", err)
			}
		}
	} else {
		if _, err := b.Discord.Rest().SetGlobalCommands(b.Discord.ApplicationID(), commands); err != nil {
			return fmt.Errorf("failed to update global commands: %w", err)
		}
	}

	var wg sync.WaitGroup
	for i := range b.Config.Nodes {
		wg.Add(1)
		config := b.Config.Nodes[i]
		go func() {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			_, err := b.Lavalink.AddNode(ctx, config)
			if err != nil {
				b.Logger.Error("failed to add node:", err)
				return
			}
		}()
	}

	return b.Discord.OpenGateway(nil)
}

func (b *Bot) OnEvent(event bot.Event) {
	switch e := event.(type) {
	case *events.VoiceServerUpdate:
		b.Logger.Debug("received voice server update")
		if e.Endpoint == nil {
			return
		}
		b.Lavalink.OnVoiceServerUpdate(context.Background(), e.GuildID, e.Token, *e.Endpoint)
	case *events.GuildVoiceStateUpdate:
		b.Logger.Debug("received voice state update")
		b.Lavalink.OnVoiceStateUpdate(context.Background(), e.VoiceState.GuildID, e.VoiceState.ChannelID, e.VoiceState.SessionID)
	}
}

func (b *Bot) NewNATSGateway() gateway.Gateway {
	return &NATSGateway{
		logger: b.Logger,
		bot:    b,
	}
}

func (b *Bot) Close() {
	_ = b.Nats.Drain()
	b.Nats.Close()
	b.Lavalink.Close()
	b.Discord.Close(context.Background())
	_ = b.Database.Close()
}
