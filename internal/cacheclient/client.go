package cacheclient

import (
	"encoding/json"
	"fmt"
	"io"
	"iter"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/snowflake/v2"
	"github.com/topi314/tint"

	"github.com/kittybot-org/kittybotgo/internal/otel"
)

// Client reads the gateway cache over HTTP. Writes are ignored because the gateway already applied the event.
type Client struct {
	cache.Caches
	base string
	http *http.Client
}

func New(baseURL string) *Client {
	return &Client{
		Caches: cache.New(cache.WithCaches(cache.FlagGuilds, cache.FlagChannels, cache.FlagRoles, cache.FlagVoiceStates)),
		base:   strings.TrimRight(baseURL, "/"),
		http: &http.Client{
			Timeout:   5 * time.Second,
			Transport: otel.Transport(nil),
		},
	}
}

func (c *Client) SelfUser() (discord.OAuth2User, bool) {
	var user discord.OAuth2User
	if !c.get("/api/self", &user) {
		return discord.OAuth2User{}, false
	}
	return user, true
}

func (c *Client) SetSelfUser(discord.OAuth2User) {}

func (c *Client) Guild(guildID snowflake.ID) (discord.CacheGuild, bool) {
	var guild discord.CacheGuild
	if !c.get("/api/guilds/"+guildID.String(), &guild) {
		return discord.CacheGuild{}, false
	}
	return guild, true
}

func (c *Client) Guilds() iter.Seq[discord.CacheGuild] {
	var guilds []discord.CacheGuild
	if !c.get("/api/guilds", &guilds) {
		guilds = nil
	}
	return func(yield func(discord.CacheGuild) bool) {
		for _, guild := range guilds {
			if !yield(guild) {
				return
			}
		}
	}
}

func (c *Client) AddGuild(discord.CacheGuild) {}

func (c *Client) RemoveGuild(snowflake.ID) (discord.CacheGuild, bool) {
	return discord.CacheGuild{}, false
}

func (c *Client) Channel(channelID snowflake.ID) (discord.GuildChannel, bool) {
	response, ok := c.getBytes("/api/channels/" + channelID.String())
	if !ok {
		return nil, false
	}
	var channel discord.UnmarshalChannel
	if err := json.Unmarshal(response, &channel); err != nil {
		slog.Error("failed to decode channel", tint.Err(err))
		return nil, false
	}
	guildChannel, ok := channel.Channel.(discord.GuildChannel)
	if !ok {
		return nil, false
	}
	return guildChannel, true
}

func (c *Client) AddChannel(discord.GuildChannel) {}

func (c *Client) RemoveChannel(snowflake.ID) (discord.GuildChannel, bool) {
	return nil, false
}

func (c *Client) RemoveChannelsByGuildID(snowflake.ID) {}

func (c *Client) AddRole(discord.Role) {}

func (c *Client) RemoveRole(snowflake.ID, snowflake.ID) (discord.Role, bool) {
	return discord.Role{}, false
}

func (c *Client) RemoveRolesByGuildID(snowflake.ID) {}

func (c *Client) AddMember(discord.Member) {}

func (c *Client) RemoveMember(snowflake.ID, snowflake.ID) (discord.Member, bool) {
	return discord.Member{}, false
}

func (c *Client) RemoveMembersByGuildID(snowflake.ID) {}

func (c *Client) VoiceState(guildID snowflake.ID, userID snowflake.ID) (discord.VoiceState, bool) {
	var state discord.VoiceState
	if !c.get(fmt.Sprintf("/api/guilds/%s/voice-states/%s", guildID, userID), &state) {
		return discord.VoiceState{}, false
	}
	return state, true
}

func (c *Client) AddVoiceState(discord.VoiceState) {}

func (c *Client) RemoveVoiceState(snowflake.ID, snowflake.ID) (discord.VoiceState, bool) {
	return discord.VoiceState{}, false
}

func (c *Client) RemoveVoiceStatesByGuildID(snowflake.ID) {}

func (c *Client) MemberPermissionsInChannel(channel discord.GuildChannel, member discord.Member) discord.Permissions {
	if channel == nil || member.User.ID == 0 {
		return discord.PermissionsNone
	}
	var body struct {
		Permissions discord.Permissions `json:"permissions"`
	}
	path := fmt.Sprintf("/api/guilds/%s/channels/%s/permissions/%s", channel.GuildID(), channel.ID(), member.User.ID)
	if !c.get(path, &body) {
		return discord.PermissionsNone
	}
	return body.Permissions
}

func (c *Client) get(path string, dest any) bool {
	body, ok := c.getBytes(path)
	if !ok {
		return false
	}
	if err := json.Unmarshal(body, dest); err != nil {
		slog.Error("failed to decode cache response", slog.String("path", path), tint.Err(err))
		return false
	}
	return true
}

func (c *Client) getBytes(path string) ([]byte, bool) {
	response, err := c.http.Get(c.base + path)
	if err != nil {
		slog.Error("cache request failed", slog.String("path", path), tint.Err(err))
		return nil, false
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusNotFound {
		return nil, false
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		slog.Error("failed to read cache response", slog.String("path", path), tint.Err(err))
		return nil, false
	}
	if response.StatusCode != http.StatusOK {
		slog.Error("cache request failed", slog.String("path", path), slog.Int("status", response.StatusCode), slog.String("body", string(body)))
		return nil, false
	}
	return body, true
}
