package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/snowflake/v2"
	"github.com/nats-io/nats.go"
	"github.com/topi314/tint"
	"go.opentelemetry.io/otel/trace"

	"github.com/kittybot-org/kittybotgo/internal/gatewaywire"
	"github.com/kittybot-org/kittybotgo/internal/otel"
)

// remoteGateway is a gateway.Gateway that does not open a Discord websocket.
// Several bot processes can share the one gateway process.
type remoteGateway struct {
	nc     *nats.Conn
	client *bot.Client
	sub    *nats.Subscription

	mu     sync.Mutex
	status gateway.Status
}

func newRemoteGateway(natsURL string) (*remoteGateway, error) {
	if natsURL == "" {
		return nil, fmt.Errorf("nats url is required")
	}
	nc, err := nats.Connect(natsURL, nats.Name("kittybot"))
	if err != nil {
		return nil, fmt.Errorf("connect nats: %w", err)
	}
	return &remoteGateway{nc: nc, status: gateway.StatusUnconnected}, nil
}

func (g *remoteGateway) bind(client *bot.Client) {
	g.client = client
}

func (g *remoteGateway) ShardID() int { return 0 }

func (g *remoteGateway) ShardCount() int { return 0 }

func (g *remoteGateway) SessionID() *string { return nil }

func (g *remoteGateway) LastSequenceReceived() *int { return nil }

func (g *remoteGateway) ResumeURL() *string { return nil }

func (g *remoteGateway) Intents() gateway.Intents {
	return gateway.IntentGuilds | gateway.IntentGuildVoiceStates | gateway.IntentGuildMembers | gateway.IntentGuildInvites
}

func (g *remoteGateway) Open(context.Context) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.sub != nil {
		g.status = gateway.StatusReady
		return nil
	}
	sub, err := g.nc.Subscribe(gatewaywire.SubjectEvents, g.onEvent)
	if err != nil {
		g.status = gateway.StatusDisconnected
		return fmt.Errorf("subscribe gateway events: %w", err)
	}
	g.sub = sub
	g.status = gateway.StatusReady
	return nil
}

func (g *remoteGateway) Close(context.Context) {
	g.CloseWithCode(context.Background(), 0, "")
}

func (g *remoteGateway) CloseWithCode(context.Context, int, string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.sub != nil {
		_ = g.sub.Unsubscribe()
		g.sub = nil
	}
	g.status = gateway.StatusDisconnected
	if g.nc != nil {
		_ = g.nc.Drain()
		g.nc = nil
	}
}

func (g *remoteGateway) Status() gateway.Status {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.status
}

func (g *remoteGateway) Send(ctx context.Context, op gateway.Opcode, data gateway.MessageData) (err error) {
	ctx, span := otel.Tracer().Start(ctx, "gateway.send", trace.WithSpanKind(trace.SpanKindClient))
	defer func() { otel.End(span, err) }()

	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}
	var guild struct {
		GuildID snowflake.ID `json:"guild_id"`
	}
	_ = json.Unmarshal(payload, &guild)
	body, err := json.Marshal(gatewaywire.SendRequest{
		Opcode:  op,
		GuildID: guild.GuildID,
		Data:    payload,
	})
	if err != nil {
		return err
	}
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}
	g.mu.Lock()
	nc := g.nc
	g.mu.Unlock()
	if nc == nil {
		return fmt.Errorf("gateway is closed")
	}
	msg := &nats.Msg{Subject: gatewaywire.SubjectSend, Data: body, Header: nats.Header{}}
	otel.Inject(ctx, msg.Header)
	reply, err := nc.RequestMsgWithContext(ctx, msg)
	if err != nil {
		return err
	}
	var response gatewaywire.SendResponse
	if err = json.Unmarshal(reply.Data, &response); err != nil {
		return err
	}
	if response.Error != "" {
		return fmt.Errorf("%s", response.Error)
	}
	return nil
}

func (g *remoteGateway) Latency() time.Duration { return 0 }

func (g *remoteGateway) Presence() *gateway.MessageDataPresenceUpdate { return nil }

func (g *remoteGateway) Shards(ctx context.Context) (shards []gatewaywire.ShardStatus, err error) {
	ctx, span := otel.Tracer().Start(ctx, "gateway.shards", trace.WithSpanKind(trace.SpanKindClient))
	defer func() { otel.End(span, err) }()

	g.mu.Lock()
	nc := g.nc
	g.mu.Unlock()
	if nc == nil {
		return nil, fmt.Errorf("gateway is closed")
	}
	msg := &nats.Msg{Subject: gatewaywire.SubjectShards, Header: nats.Header{}}
	otel.Inject(ctx, msg.Header)
	reply, err := nc.RequestMsgWithContext(ctx, msg)
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal(reply.Data, &shards); err != nil {
		return nil, err
	}
	return shards, nil
}

func (g *remoteGateway) onEvent(msg *nats.Msg) {
	var event gatewaywire.Event
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		slog.Error("failed to decode gateway event", tint.Err(err))
		return
	}
	data, err := gateway.UnmarshalEventData(event.Data, event.Type)
	if err != nil {
		slog.Error("failed to unmarshal gateway event", slog.String("type", string(event.Type)), tint.Err(err))
		return
	}
	if g.client == nil {
		return
	}
	g.client.EventManager.HandleGatewayEvent(g, event.Type, event.Sequence, data)
}
