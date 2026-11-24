package gateway

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"slices"
	"time"

	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/disgo/sharding"
	"github.com/nats-io/nats.go"
	"github.com/topi314/tint"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/kittybot-org/kittybotgo/internal/gatewaywire"
	"github.com/kittybot-org/kittybotgo/internal/otel"
)

// Gateway owns the Discord shards, the memory cache, and the HTTP cache API.
type Gateway struct {
	Config Config
	cache  *Cache
	rest   rest.Rest
	nats   *nats.Conn
	shards sharding.ShardManager
	http   *http.Server
	subs   []*nats.Subscription
}

func New(cfg Config) *Gateway {
	return &Gateway{
		Config: cfg,
		cache:  NewCache(),
	}
}

func (g *Gateway) Start() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if g.Config.Gateway.Token == "" {
		return fmt.Errorf("bot token is required")
	}
	if g.Config.Gateway.NATSURL == "" {
		return fmt.Errorf("nats url is required")
	}

	var restOpts []rest.ClientConfigOpt
	if g.Config.Gateway.RestURL != "" {
		restOpts = append(restOpts,
			rest.WithURL(g.Config.Gateway.RestURL),
			rest.WithRateLimiter(rest.NewNoopRateLimiter()),
		)
	}
	restOpts = append(restOpts, rest.WithHTTPClient(&http.Client{
		Transport: otel.Transport(nil),
	}))
	g.rest = rest.New(rest.NewClient(g.Config.Gateway.Token, restOpts...))

	nc, err := nats.Connect(g.Config.Gateway.NATSURL, nats.Name("kittybot-gateway"))
	if err != nil {
		return fmt.Errorf("connect nats: %w", err)
	}
	g.nats = nc

	sendSub, err := nc.Subscribe(gatewaywire.SubjectSend, g.onSend)
	if err != nil {
		return fmt.Errorf("subscribe send: %w", err)
	}
	shardSub, err := nc.Subscribe(gatewaywire.SubjectShards, g.onShards)
	if err != nil {
		return fmt.Errorf("subscribe shards: %w", err)
	}
	g.subs = []*nats.Subscription{sendSub, shardSub}

	g.http = &http.Server{
		Addr:    g.Config.Gateway.Address(),
		Handler: otel.Handler(g.routes(), "cache"),
	}
	go func() {
		slog.Info("cache listening", slog.String("addr", g.http.Addr))
		if err := g.http.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("cache server stopped", tint.Err(err))
		}
	}()

	gatewayBot, err := g.rest.GetGatewayBot()
	if err != nil {
		return fmt.Errorf("get gateway bot: %w", err)
	}
	shardCount := max(gatewayBot.Shards, 1)
	shardIDs := make([]int, shardCount)
	for i := range shardIDs {
		shardIDs[i] = i
	}
	concurrency := max(gatewayBot.SessionStartLimit.MaxConcurrency, 1)

	gatewayOpts := []gateway.ConfigOpt{
		gateway.WithEnableRawEvents(true),
		gateway.WithIntents(
			gateway.IntentGuilds,
			gateway.IntentGuildVoiceStates,
			gateway.IntentGuildMembers,
			gateway.IntentGuildInvites,
		),
	}
	gatewayURL := gatewayBot.URL
	if g.Config.Gateway.GatewayURL != "" {
		gatewayURL = g.Config.Gateway.GatewayURL
		gatewayOpts = append(gatewayOpts, gateway.WithCompression(gateway.CompressionNone))
	}
	gatewayOpts = append(gatewayOpts, gateway.WithURL(gatewayURL))

	g.shards = sharding.New(g.Config.Gateway.Token, g.onEvent,
		sharding.WithShardCount(shardCount),
		sharding.WithShardIDs(shardIDs...),
		sharding.WithGatewayConfigOpts(gatewayOpts...),
		sharding.WithIdentifyRateLimiterConfigOpt(gateway.WithIdentifyMaxConcurrency(concurrency)),
	)

	slog.Info("opening shards", slog.Int("count", shardCount), slog.String("url", gatewayURL))
	g.shards.Open(ctx)
	return nil
}

func (g *Gateway) Close() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if g.shards != nil {
		g.shards.Close(ctx)
	}
	if g.http != nil {
		_ = g.http.Shutdown(ctx)
	}
	for _, sub := range g.subs {
		_ = sub.Unsubscribe()
	}
	if g.nats != nil {
		_ = g.nats.Drain()
	}
	if g.rest != nil {
		g.rest.Close(ctx)
	}
}

func (g *Gateway) onEvent(shard gateway.Gateway, eventType gateway.EventType, sequence int, event gateway.EventData) {
	if eventType == gateway.EventTypeRaw {
		g.forwardRaw(shard, sequence, event)
		return
	}
	g.cache.Handle(event)
}

func (g *Gateway) forwardRaw(shard gateway.Gateway, sequence int, event gateway.EventData) {
	raw, ok := event.(gateway.EventRaw)
	if !ok {
		return
	}
	payload, err := io.ReadAll(raw.Payload)
	if err != nil {
		slog.Error("failed to read raw gateway event", tint.Err(err))
		return
	}
	body, err := json.Marshal(gatewaywire.Event{
		Type:     raw.EventType,
		Sequence: sequence,
		ShardID:  shard.ShardID(),
		Data:     payload,
	})
	if err != nil {
		slog.Error("failed to encode gateway event", tint.Err(err))
		return
	}
	if err = g.nats.Publish(gatewaywire.SubjectEvents, body); err != nil {
		slog.Error("failed to publish gateway event", slog.String("type", string(raw.EventType)), tint.Err(err))
	}
}

func (g *Gateway) onSend(msg *nats.Msg) {
	ctx := otel.Extract(context.Background(), msg.Header)
	ctx, span := otel.Tracer().Start(ctx, "gateway.send", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	var request gatewaywire.SendRequest
	if err := json.Unmarshal(msg.Data, &request); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		g.reply(msg, gatewaywire.SendResponse{Error: err.Error()})
		return
	}
	shard := g.shards.ShardByGuildID(request.GuildID)
	if shard == nil {
		err := fmt.Errorf("shard not found")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		g.reply(msg, gatewaywire.SendResponse{Error: err.Error()})
		return
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := shard.Send(ctx, request.Opcode, gateway.MessageDataUnknown(request.Data)); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		g.reply(msg, gatewaywire.SendResponse{Error: err.Error()})
		return
	}
	g.reply(msg, gatewaywire.SendResponse{})
}

func (g *Gateway) onShards(msg *nats.Msg) {
	ctx := otel.Extract(context.Background(), msg.Header)
	_, span := otel.Tracer().Start(ctx, "gateway.shards", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	statuses := []gatewaywire.ShardStatus{}
	if g.shards != nil {
		for shard := range g.shards.Shards() {
			session := ""
			if id := shard.SessionID(); id != nil {
				session = *id
			}
			statuses = append(statuses, gatewaywire.ShardStatus{
				ID:        shard.ShardID(),
				Status:    shard.Status().String(),
				LatencyMS: shard.Latency().Milliseconds(),
				SessionID: session,
			})
		}
	}
	slices.SortFunc(statuses, func(a, b gatewaywire.ShardStatus) int {
		return a.ID - b.ID
	})
	body, err := json.Marshal(statuses)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		slog.Error("failed to encode shard status", tint.Err(err))
		return
	}
	if err = msg.Respond(body); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		slog.Error("failed to respond with shard status", tint.Err(err))
	}
}

func (g *Gateway) reply(msg *nats.Msg, response gatewaywire.SendResponse) {
	body, err := json.Marshal(response)
	if err != nil {
		return
	}
	if err = msg.Respond(body); err != nil {
		slog.Error("failed to respond to gateway send", tint.Err(err))
	}
}
