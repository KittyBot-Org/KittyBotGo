package gateway

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/snowflake/v2"
	"github.com/topi314/tint"
)

func (g *Gateway) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/self", g.getSelf)
	mux.HandleFunc("GET /api/guilds", g.getGuilds)
	mux.HandleFunc("GET /api/guilds/{id}", g.getGuild)
	mux.HandleFunc("GET /api/channels/{channelId}", g.getChannel)
	mux.HandleFunc("GET /api/guilds/{id}/channels/{channelId}", g.getGuildChannel)
	mux.HandleFunc("GET /api/guilds/{id}/voice-states/{userId}", g.getVoiceState)
	mux.HandleFunc("GET /api/guilds/{id}/channels/{channelId}/permissions/{userId}", g.getPermissions)
	return mux
}

func (g *Gateway) getSelf(w http.ResponseWriter, r *http.Request) {
	user, ok := g.cache.SelfUser()
	if !ok {
		http.NotFound(w, r)
		return
	}
	writeJSON(w, user)
}

func (g *Gateway) getGuilds(w http.ResponseWriter, _ *http.Request) {
	guilds := make([]discord.CacheGuild, 0)
	for guild := range g.cache.Guilds() {
		guilds = append(guilds, guild)
	}
	writeJSON(w, guilds)
}

func (g *Gateway) getGuild(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	guild, found := g.cache.Guild(id)
	if !found {
		http.NotFound(w, r)
		return
	}
	writeJSON(w, guild)
}

func (g *Gateway) getChannel(w http.ResponseWriter, r *http.Request) {
	channelID, ok := pathID(w, r, "channelId")
	if !ok {
		return
	}
	channel, found := g.cache.Channel(channelID)
	if !found {
		http.NotFound(w, r)
		return
	}
	writeJSON(w, channel)
}

func (g *Gateway) getGuildChannel(w http.ResponseWriter, r *http.Request) {
	guildID, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	channelID, ok := pathID(w, r, "channelId")
	if !ok {
		return
	}
	channel, found := g.cache.Channel(channelID)
	if !found || channel.GuildID() != guildID {
		http.NotFound(w, r)
		return
	}
	writeJSON(w, channel)
}

func (g *Gateway) getVoiceState(w http.ResponseWriter, r *http.Request) {
	guildID, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	userID, ok := pathID(w, r, "userId")
	if !ok {
		return
	}
	state, found := g.cache.VoiceState(guildID, userID)
	if !found {
		http.NotFound(w, r)
		return
	}
	writeJSON(w, state)
}

func (g *Gateway) getPermissions(w http.ResponseWriter, r *http.Request) {
	guildID, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	channelID, ok := pathID(w, r, "channelId")
	if !ok {
		return
	}
	userID, ok := pathID(w, r, "userId")
	if !ok {
		return
	}
	channel, found := g.cache.Channel(channelID)
	if !found || channel.GuildID() != guildID {
		http.NotFound(w, r)
		return
	}
	member, found := g.member(guildID, userID)
	if !found {
		http.NotFound(w, r)
		return
	}
	writeJSON(w, map[string]discord.Permissions{
		"permissions": g.cache.MemberPermissionsInChannel(channel, member),
	})
}

func (g *Gateway) member(guildID snowflake.ID, userID snowflake.ID) (discord.Member, bool) {
	member, ok := g.cache.Member(guildID, userID)
	if ok || g.rest == nil {
		return member, ok
	}
	fetched, err := g.rest.GetMember(guildID, userID)
	if err != nil {
		return discord.Member{}, false
	}
	fetched.GuildID = guildID
	g.cache.AddMember(*fetched)
	return g.cache.Member(guildID, userID)
}

func pathID(w http.ResponseWriter, r *http.Request, name string) (snowflake.ID, bool) {
	id, err := snowflake.Parse(r.PathValue(name))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return 0, false
	}
	return id, true
}

func writeJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(value); err != nil {
		slog.Error("failed to write cache response", tint.Err(err))
	}
}
