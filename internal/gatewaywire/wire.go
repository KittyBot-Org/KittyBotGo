package gatewaywire

import (
	"encoding/json"

	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/snowflake/v2"
)

const (
	SubjectEvents = "kittybot.gateway.events"
	SubjectSend   = "kittybot.gateway.send"
	SubjectShards = "kittybot.gateway.shards"
)

// Event is one Discord dispatch, published to every bot session.
type Event struct {
	Type     gateway.EventType `json:"type"`
	Sequence int               `json:"sequence"`
	ShardID  int               `json:"shard_id"`
	Data     json.RawMessage   `json:"data"`
}

// SendRequest asks the gateway process to send an opcode on the shard for GuildID.
type SendRequest struct {
	Opcode  gateway.Opcode  `json:"opcode"`
	GuildID snowflake.ID    `json:"guild_id"`
	Data    json.RawMessage `json:"data"`
}

// SendResponse is the reply to a SendRequest.
type SendResponse struct {
	Error string `json:"error,omitempty"`
}

// ShardStatus is one shard owned by the gateway process.
type ShardStatus struct {
	ID        int    `json:"id"`
	Status    string `json:"status"`
	LatencyMS int64  `json:"latency_ms"`
	SessionID string `json:"session_id"`
}
