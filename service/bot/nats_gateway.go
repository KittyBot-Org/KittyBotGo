package bot

import (
	"context"
	"fmt"
	"time"

	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/json"
	"github.com/disgoorg/log"
	"github.com/nats-io/nats.go"
)

type shardIDKey struct{}

var shardIDKeyVal = shardIDKey{}

func ShardIDFromContext(ctx context.Context) (int, bool) {
	shardID, ok := ctx.Value(shardIDKeyVal).(int)
	return shardID, ok
}

func ContextWithShardID(ctx context.Context, shardID int) context.Context {
	return context.WithValue(ctx, shardIDKeyVal, shardID)
}

var _ gateway.Gateway = (*NATSGateway)(nil)

type NATSGateway struct {
	logger log.Logger
	bot    *Bot
}

func (n *NATSGateway) ShardID() int {
	return 0
}

func (n *NATSGateway) ShardCount() int {
	return 1
}

func (n *NATSGateway) SessionID() *string {
	return json.Ptr("session-id")
}

func (n *NATSGateway) LastSequenceReceived() *int {
	return json.Ptr(0)
}

func (n *NATSGateway) Intents() gateway.Intents {
	return gateway.IntentsAll
}

func (n *NATSGateway) Open(ctx context.Context) error {
	_, err := n.bot.Nats.QueueSubscribe("gateway.*.events.*", n.bot.Config.Nats.Queue, func(msg *nats.Msg) {
		eventType := gateway.EventType(msg.Subject[len("gateway.*.events."):])
		n.logger.Debugf("received event: %s\ndata: %s", eventType, msg.Data)

		data, err := gateway.UnmarshalEventData(msg.Data, eventType)
		if err != nil {
			n.logger.Errorf("failed to unmarshal data: %s", err)
			return
		}

		n.bot.Discord.EventManager().HandleGatewayEvent(eventType, 0, 0, data)
	})
	return err
}

func (n *NATSGateway) Close(ctx context.Context) {
	n.CloseWithCode(ctx, 1000, "Normal Closure")
}

func (n *NATSGateway) CloseWithCode(_ context.Context, code int, message string) {
	n.logger.Infof("closing nats gateway with code %d and message %s", code, message)
	_ = n.bot.Nats.Drain()
	n.bot.Nats.Close()
}

func (n *NATSGateway) Status() gateway.Status {
	return gateway.StatusReady
}

func (n *NATSGateway) Send(ctx context.Context, op gateway.Opcode, data gateway.MessageData) error {
	shardID, ok := ShardIDFromContext(ctx)
	if !ok {
		return fmt.Errorf("no shard id found in context")
	}

	raw, err := json.Marshal(gateway.Message{
		Op: op,
		D:  data,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	return n.bot.Nats.Publish(fmt.Sprintf("gateway.%d.commands", shardID), raw)
}

func (n *NATSGateway) Latency() time.Duration {
	return time.Millisecond
}

func (n *NATSGateway) Presence() *gateway.MessageDataPresenceUpdate {
	return nil
}
