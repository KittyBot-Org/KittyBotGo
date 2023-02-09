package gateway

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/disgo/sharding"
	"github.com/disgoorg/json"
	"github.com/disgoorg/log"
	"github.com/nats-io/nats.go"
)

func New(logger log.Logger, cfg Config) (*Gateway, error) {
	shardCount := cfg.ShardCount
	if shardCount == 0 {
		discordRest := rest.New(rest.NewClient(cfg.Token, rest.WithLogger(logger)))
		gatewayBot, err := discordRest.GetGatewayBot()
		if err != nil {
			return nil, fmt.Errorf("failed to get gateway bot: %w", err)
		}
		shardCount = gatewayBot.Shards
	}

	g := &Gateway{
		Config: cfg,
		Logger: logger,
	}

	shardIDs := make([]int, shardCount)
	for i := 0; i < shardCount; i++ {
		shardIDs[i] = i
	}

	discord := sharding.New(cfg.Token,
		g.OnEvent,
		sharding.WithLogger(logger),
		sharding.WithAutoScaling(true),
		sharding.WithShardIDs(shardIDs...),
		sharding.WithShardCount(shardCount),
		sharding.WithGatewayConfigOpts(
			gateway.WithLogger(logger),
			gateway.WithIntents(
				gateway.IntentGuilds,
				gateway.IntentGuildMembers,
				gateway.IntentGuildPresences,
				gateway.IntentGuildVoiceStates,
				gateway.IntentGuildInvites,
				gateway.IntentGuildModeration,
				gateway.IntentAutoModerationExecution,
			),
		),
	)

	conn, err := nats.Connect(cfg.Nats.URL,
		nats.Name("gateway"),
		nats.UserInfo(cfg.Nats.User, cfg.Nats.Password),
		nats.MaxReconnects(-1),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to nats: %w", err)
	}

	g.Discord = discord
	g.Nats = conn
	return g, nil
}

type Gateway struct {
	Config  Config
	Logger  log.Logger
	Discord sharding.ShardManager
	Nats    *nats.Conn
}

func (g *Gateway) Start(ctx context.Context) error {
	g.Discord.Open(ctx)

	if _, err := g.Nats.Subscribe("gateway.*.commands", func(msg *nats.Msg) {
		shardID, _ := strconv.Atoi(msg.Subject[len("gateway."):strings.Index(msg.Subject, ".commands")])

		var cmd gateway.Message
		if err := json.Unmarshal(msg.Data, &cmd); err != nil {
			g.Logger.Errorf("Failed to unmarshal command: %v", err)
			return
		}
		
		if err := g.Discord.Shard(shardID).Send(context.Background(), cmd.Op, cmd.D); err != nil {
			g.Logger.Errorf("Failed to send command: %v", err)
		}
	}); err != nil {
		return fmt.Errorf("failed to subscribe to nats: %w", err)
	}

	return nil
}

func (g *Gateway) OnEvent(eventType gateway.EventType, _ int, shardID int, event gateway.EventData) {
	switch eventType {
	case gateway.EventTypeReady:
		g.Logger.Debugf("Shard [%d/%d] is ready", shardID, len(g.Discord.Shards()))
		return

	case gateway.EventTypeResumed:
		g.Logger.Debugf("Shard [%d/%d] is  resumed", shardID, len(g.Discord.Shards()))
		return
	}

	data, err := json.Marshal(event)
	if err != nil {
		g.Logger.Errorf("Failed to marshal event: %v", err)
		return
	}

	if err = g.Nats.Publish(fmt.Sprintf("gateway.%d.events.%s", shardID, eventType), data); err != nil {
		g.Logger.Errorf("Failed to publish event: %v", err)
	}
}

func (g *Gateway) Close(ctx context.Context) {
	g.Discord.Close(ctx)
	_ = g.Nats.Drain()
	g.Nats.Close()
}
