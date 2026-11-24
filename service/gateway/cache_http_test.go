package gateway

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/snowflake/v2"

	"github.com/kittybot-org/kittybotgo/internal/cacheclient"
)

func TestCacheHTTP(t *testing.T) {
	cache := NewCache()
	handle(t, cache, gateway.EventTypeReady, []byte(`{
		"v": 10,
		"user": {"id": "300", "username": "bot", "discriminator": "0"},
		"guilds": [],
		"session_id": "session",
		"resume_gateway_url": "wss://example",
		"application": {"id": "300"}
	}`))
	handle(t, cache, gateway.EventTypeGuildCreate, []byte(`{
		"id": "200",
		"name": "guild",
		"owner_id": "300",
		"roles": [{"id": "200", "name": "@everyone", "permissions": "1024"}],
		"channels": [{"id": "100", "type": 2, "name": "voice", "guild_id": "200"}],
		"members": [{"user": {"id": "300", "username": "bot", "discriminator": "0"}, "roles": ["200"], "joined_at": "2020-01-01T00:00:00.000000+00:00"}],
		"voice_states": [{"guild_id": "200", "channel_id": "100", "user_id": "300", "session_id": "abc", "deaf": false, "mute": false, "self_deaf": false, "self_mute": false, "suppress": false}]
	}`))

	server := httptest.NewServer((&Gateway{cache: cache}).routes())
	defer server.Close()

	client := cacheclient.New(server.URL)
	user, ok := client.SelfUser()
	if !ok || user.ID != snowflake.ID(300) {
		t.Fatalf("self user = %#v, %t", user, ok)
	}

	guilds := 0
	for guild := range client.Guilds() {
		guilds++
		if guild.ID != snowflake.ID(200) || guild.Name != "guild" {
			t.Fatalf("guild = %#v", guild)
		}
	}
	if guilds != 1 {
		t.Fatalf("guilds = %d", guilds)
	}

	channel, ok := client.Channel(100)
	if !ok || channel.GuildID() != 200 || channel.Name() != "voice" {
		t.Fatalf("channel = %#v, %t", channel, ok)
	}

	state, ok := client.VoiceState(200, 300)
	if !ok || state.ChannelID == nil || *state.ChannelID != 100 {
		t.Fatalf("voice state = %#v, %t", state, ok)
	}

	permissions := client.MemberPermissionsInChannel(channel, discord.Member{
		User:    discord.User{ID: 300},
		GuildID: 200,
	})
	if permissions == 0 {
		t.Fatal("expected permissions")
	}

	response, err := http.Get(server.URL + "/api/guilds/200/channels/100")
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("guild channel status = %d", response.StatusCode)
	}
}

func handle(t *testing.T, cache *Cache, eventType gateway.EventType, payload []byte) {
	t.Helper()
	event, err := gateway.UnmarshalEventData(payload, eventType)
	if err != nil {
		t.Fatal(err)
	}
	cache.Handle(event)
}
