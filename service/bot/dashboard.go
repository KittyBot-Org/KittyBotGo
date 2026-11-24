package bot

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/snowflake/v2"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/topi314/tint"

	"github.com/kittybot-org/kittybotgo/internal/gatewaywire"
	"github.com/kittybot-org/kittybotgo/internal/otel"
	"github.com/kittybot-org/kittybotgo/service/bot/db/dbsqlc"
)

func (b *Bot) startDashboard() {
	if b.Config.Dashboard.Port <= 0 {
		return
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health_check", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("alive"))
	})
	mux.HandleFunc("GET /info", b.dashboardInfo)
	mux.HandleFunc("GET /shards", b.dashboardShards)
	mux.HandleFunc("GET /discord_login", b.dashboardDiscordLogin)
	mux.HandleFunc("GET /bot_invite", b.dashboardBotInvite)
	mux.HandleFunc("GET /discord_invite", b.dashboardBotInvite)
	mux.HandleFunc("GET /commands", b.dashboardCommands)
	mux.HandleFunc("POST /login", b.dashboardLogin)
	mux.HandleFunc("DELETE /login", b.dashboardLogout)
	mux.HandleFunc("GET /user/me", b.dashboardMe)
	mux.HandleFunc("POST /webhooks/votes/{botlist}", b.dashboardVote)
	mux.HandleFunc("GET /guilds", b.dashboardGuilds)
	mux.HandleFunc("GET /guilds/{guildId}/roles", b.dashboardRoles)
	mux.HandleFunc("GET /guilds/{guildId}/channels", b.dashboardChannels)
	mux.HandleFunc("GET /guilds/{guildId}/emotes", b.dashboardEmotes)
	mux.HandleFunc("GET /guilds/{guildId}/users", b.dashboardUsers)
	mux.HandleFunc("GET /guilds/{guildId}/invites", b.dashboardInvites)
	mux.HandleFunc("GET /guilds/{guildId}/tags", b.dashboardTags)
	mux.HandleFunc("POST /guilds/{guildId}/tags/{tagId}", b.dashboardPostTag)
	mux.HandleFunc("DELETE /guilds/{guildId}/tags/{tagId}", b.dashboardDeleteTag)
	mux.HandleFunc("GET /guilds/{guildId}/settings", b.dashboardGetSettings)
	mux.HandleFunc("POST /guilds/{guildId}/settings", b.dashboardPostSettings)

	b.httpServer = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", b.Config.Dashboard.Host, b.Config.Dashboard.Port),
		Handler: otel.Handler(b.withCORS(mux), "dashboard"),
	}
	go func() {
		slog.Info("dashboard listening", slog.String("addr", b.httpServer.Addr))
		if err := b.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("dashboard stopped", tint.Err(err))
		}
	}()
}

func (b *Bot) withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := b.Config.Dashboard.OriginURL
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
		}
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (b *Bot) dashboardInfo(w http.ResponseWriter, _ *http.Request) {
	user, _ := b.Discord.Caches.SelfUser()
	guilds := 0
	for range b.Discord.Caches.Guilds() {
		guilds++
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"id":      user.ID.String(),
		"name":    user.Username,
		"version": b.Version,
		"commit":  b.Commit,
		"guilds":  guilds,
		"uptime":  time.Since(b.Started).String(),
	})
}

func (b *Bot) dashboardShards(w http.ResponseWriter, r *http.Request) {
	shards, err := b.remote.Shards(r.Context())
	if err != nil {
		slog.Error("failed to load shard status", tint.Err(err))
		http.Error(w, "failed to load shards", http.StatusBadGateway)
		return
	}
	if shards == nil {
		shards = []gatewaywire.ShardStatus{}
	}
	writeJSON(w, http.StatusOK, shards)
}

func (b *Bot) dashboardDiscordLogin(w http.ResponseWriter, r *http.Request) {
	if _, ok := b.authorizedUser(r); ok {
		http.Redirect(w, r, b.Config.Dashboard.RedirectURL, http.StatusFound)
		return
	}
	state := randomToken()
	b.stateMu.Lock()
	b.states[state] = time.Now().Add(10 * time.Minute)
	b.stateMu.Unlock()
	http.Redirect(w, r, discord.AuthorizeURL(discord.QueryValues{
		"client_id":     b.Discord.ApplicationID.String(),
		"redirect_uri":  b.Config.Dashboard.RedirectURL,
		"response_type": "code",
		"scope":         string(discord.OAuth2ScopeIdentify) + " " + string(discord.OAuth2ScopeGuilds),
		"state":         state,
	}), http.StatusFound)
}

func (b *Bot) dashboardBotInvite(w http.ResponseWriter, r *http.Request) {
	target := b.Config.Dashboard.BotInviteURL
	if target == "" {
		target = discord.AuthorizeURL(discord.QueryValues{
			"client_id": b.Discord.ApplicationID.String(),
			"scope":     "bot applications.commands",
		})
	}
	http.Redirect(w, r, target, http.StatusFound)
}

func (b *Bot) dashboardCommands(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, []string{"music", "queue", "history", "playlists", "liked-songs", "settings", "admin", "tags", "notification", "neko", "utility", "info", "ping"})
}

func (b *Bot) dashboardLogin(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Code  string `json:"code"`
		State string `json:"state"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Code == "" || body.State == "" {
		http.Error(w, "State or code is invalid", http.StatusUnauthorized)
		return
	}
	b.stateMu.Lock()
	expiry, ok := b.states[body.State]
	delete(b.states, body.State)
	b.stateMu.Unlock()
	if !ok || time.Now().After(expiry) {
		http.Error(w, "State invalid/expired. Please try again", http.StatusUnauthorized)
		return
	}
	token, err := b.Discord.Rest.GetAccessToken(b.Discord.ApplicationID, b.Config.Bot.Secret, body.Code, b.Config.Dashboard.RedirectURL)
	if err != nil {
		http.Error(w, "Could not login", http.StatusForbidden)
		return
	}
	user, err := b.Discord.Rest.GetCurrentUser(token.AccessToken)
	if err != nil {
		http.Error(w, "Could not login", http.StatusForbidden)
		return
	}
	if err := b.Database.UpsertSession(r.Context(), dbsqlc.UpsertSessionParams{
		UserID:       int64(user.ID),
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		Expiration:   pgtype.Timestamp{Time: time.Now().Add(token.ExpiresIn), Valid: true},
	}); err != nil {
		http.Error(w, "Could not login", http.StatusForbidden)
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]string{"token": signJWT(user.ID, b.Config.Dashboard.SigningKey)})
}

func (b *Bot) dashboardLogout(w http.ResponseWriter, r *http.Request) {
	userID, ok := b.authorizedUser(r)
	if !ok {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}
	_ = b.Database.DeleteSession(r.Context(), int64(userID))
	w.WriteHeader(http.StatusNoContent)
}

func (b *Bot) dashboardMe(w http.ResponseWriter, r *http.Request) {
	session, ok := b.authorizedSession(r)
	if !ok {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}
	user, err := b.Discord.Rest.GetCurrentUser(session.AccessToken)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"id":            user.ID.String(),
		"username":      user.Username,
		"discriminator": user.Discriminator,
		"avatar":        user.Avatar,
	})
}

func (b *Bot) dashboardVote(w http.ResponseWriter, r *http.Request) {
	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	raw, _ := body["user"].(string)
	if raw == "" {
		raw, _ = body["user_id"].(string)
	}
	userID, err := snowflake.Parse(raw)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if err := b.Database.UpsertVote(r.Context(), dbsqlc.UpsertVoteParams{
		UserID:     int64(userID),
		VoteExpiry: pgtype.Timestamp{Time: time.Now().Add(12 * time.Hour), Valid: true},
	}); err != nil {
		http.Error(w, "failed", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (b *Bot) dashboardGuilds(w http.ResponseWriter, r *http.Request) {
	session, ok := b.authorizedSession(r)
	if !ok {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}
	guilds, err := b.Discord.Rest.GetCurrentUserGuilds(session.AccessToken, 0, 0, 200, 0, false)
	if err != nil {
		http.Error(w, "failed to load guilds", http.StatusBadGateway)
		return
	}
	var visible []map[string]any
	for _, guild := range guilds {
		if !b.canManage(snowflake.ID(session.UserID), guild.ID, guild.Permissions, guild.Owner) {
			continue
		}
		visible = append(visible, map[string]any{"id": guild.ID.String(), "name": guild.Name, "icon": guild.Icon, "owner": guild.Owner})
	}
	writeJSON(w, http.StatusOK, visible)
}

func (b *Bot) dashboardRoles(w http.ResponseWriter, r *http.Request) {
	guildID, ok := b.guildParam(w, r)
	if !ok {
		return
	}
	roles, err := b.Discord.Rest.GetRoles(guildID)
	if err != nil {
		http.Error(w, "failed to load roles", http.StatusBadGateway)
		return
	}
	var payload []map[string]any
	for _, role := range roles {
		payload = append(payload, map[string]any{"id": role.ID.String(), "name": role.Name, "color": role.Color, "position": role.Position})
	}
	writeJSON(w, http.StatusOK, payload)
}

func (b *Bot) dashboardChannels(w http.ResponseWriter, r *http.Request) {
	guildID, ok := b.guildParam(w, r)
	if !ok {
		return
	}
	channels, err := b.Discord.Rest.GetGuildChannels(guildID)
	if err != nil {
		http.Error(w, "failed to load channels", http.StatusBadGateway)
		return
	}
	var payload []map[string]any
	for _, channel := range channels {
		payload = append(payload, map[string]any{"id": channel.ID().String(), "name": channel.Name(), "type": channel.Type()})
	}
	writeJSON(w, http.StatusOK, payload)
}

func (b *Bot) dashboardEmotes(w http.ResponseWriter, r *http.Request) {
	guildID, ok := b.guildParam(w, r)
	if !ok {
		return
	}
	emojis, err := b.Discord.Rest.GetEmojis(guildID)
	if err != nil {
		http.Error(w, "failed to load emotes", http.StatusBadGateway)
		return
	}
	var payload []map[string]any
	for _, emoji := range emojis {
		payload = append(payload, map[string]any{"id": emoji.ID.String(), "name": emoji.Name, "animated": emoji.Animated})
	}
	writeJSON(w, http.StatusOK, payload)
}

func (b *Bot) dashboardUsers(w http.ResponseWriter, r *http.Request) {
	guildID, ok := b.guildParam(w, r)
	if !ok {
		return
	}
	members, err := b.Discord.Rest.GetMembers(guildID, 1000, 0)
	if err != nil {
		http.Error(w, "failed to load users", http.StatusBadGateway)
		return
	}
	var payload []map[string]any
	for _, member := range members {
		payload = append(payload, map[string]any{"id": member.User.ID.String(), "username": member.User.Username, "nick": member.Nick})
	}
	writeJSON(w, http.StatusOK, payload)
}

func (b *Bot) dashboardInvites(w http.ResponseWriter, r *http.Request) {
	guildID, ok := b.guildParam(w, r)
	if !ok {
		return
	}
	invites, err := b.Discord.Rest.GetGuildInvites(guildID)
	if err != nil {
		http.Error(w, "failed to load invites", http.StatusBadGateway)
		return
	}
	var payload []map[string]any
	for _, invite := range invites {
		payload = append(payload, map[string]any{"code": invite.Code, "uses": invite.Uses, "max_uses": invite.MaxUses})
	}
	writeJSON(w, http.StatusOK, payload)
}

func (b *Bot) dashboardTags(w http.ResponseWriter, r *http.Request) {
	guildID, ok := b.guildParam(w, r)
	if !ok {
		return
	}
	tags, err := b.Database.ListTags(r.Context(), int64(guildID))
	if err != nil {
		http.Error(w, "failed to load tags", http.StatusInternalServerError)
		return
	}
	var payload []map[string]any
	for _, tag := range tags {
		payload = append(payload, tagJSON(tag))
	}
	writeJSON(w, http.StatusOK, payload)
}

func (b *Bot) dashboardPostTag(w http.ResponseWriter, r *http.Request) {
	guildID, ok := b.guildParam(w, r)
	if !ok {
		return
	}
	tagID, err := strconv.Atoi(r.PathValue("tagId"))
	if err != nil {
		http.Error(w, "bad tag", http.StatusBadRequest)
		return
	}
	var body struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	tag, err := b.Database.GetTagByID(r.Context(), dbsqlc.GetTagByIDParams{GuildID: int64(guildID), ID: int64(tagID)})
	if err != nil {
		http.Error(w, "tag not found", http.StatusNotFound)
		return
	}
	if err := b.Database.UpdateTag(r.Context(), dbsqlc.UpdateTagParams{
		GuildID: int64(guildID),
		Name:    tag.Name,
		Content: body.Content,
	}); err != nil {
		http.Error(w, "failed to update tag", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (b *Bot) dashboardDeleteTag(w http.ResponseWriter, r *http.Request) {
	guildID, ok := b.guildParam(w, r)
	if !ok {
		return
	}
	tagID, err := strconv.Atoi(r.PathValue("tagId"))
	if err != nil {
		http.Error(w, "bad tag", http.StatusBadRequest)
		return
	}
	tag, err := b.Database.GetTagByID(r.Context(), dbsqlc.GetTagByIDParams{GuildID: int64(guildID), ID: int64(tagID)})
	if err != nil {
		http.Error(w, "tag not found", http.StatusNotFound)
		return
	}
	if _, err := b.Database.DeleteTag(r.Context(), dbsqlc.DeleteTagParams{GuildID: int64(guildID), Name: tag.Name}); err != nil {
		http.Error(w, "failed to delete tag", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (b *Bot) dashboardGetSettings(w http.ResponseWriter, r *http.Request) {
	guildID, ok := b.guildParam(w, r)
	if !ok {
		return
	}
	settings, err := b.guild(r.Context(), guildID)
	if err != nil {
		http.Error(w, "failed to load settings", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"announcement_channel_id": idString(settings.AnnouncementChannelID),
		"join_messages_enabled":   settings.JoinMessagesEnabled,
		"join_message":            settings.JoinMessage,
		"leave_messages_enabled":  settings.LeaveMessagesEnabled,
		"leave_message":           settings.LeaveMessage,
		"nsfw_enabled":            settings.NsfwEnabled,
		"dj_role_id":              idString(settings.DjRoleID),
		"role_saver_enabled":      settings.RoleSaverEnabled,
	})
}

func (b *Bot) dashboardPostSettings(w http.ResponseWriter, r *http.Request) {
	guildID, ok := b.guildParam(w, r)
	if !ok {
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	fields := map[string]string{
		"announcement_channel_id": "announcement_channel_id",
		"join_message":            "join_message",
		"join_messages_enabled":   "join_messages_enabled",
		"leave_message":           "leave_message",
		"leave_messages_enabled":  "leave_messages_enabled",
		"nsfw_enabled":            "nsfw_enabled",
		"dj_role_id":              "dj_role_id",
		"role_saver_enabled":      "role_saver_enabled",
	}
	for key, column := range fields {
		value, exists := payload[key]
		if !exists {
			continue
		}
		parsed, ok := settingValue(key, value)
		if !ok {
			continue
		}
		if err := b.Database.SetGuildColumn(r.Context(), int64(guildID), column, parsed); err != nil {
			http.Error(w, "failed to save settings", http.StatusInternalServerError)
			return
		}
	}
	w.WriteHeader(http.StatusAccepted)
}

func (b *Bot) guildParam(w http.ResponseWriter, r *http.Request) (snowflake.ID, bool) {
	session, ok := b.authorizedSession(r)
	if !ok {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return 0, false
	}
	guildID, err := snowflake.Parse(r.PathValue("guildId"))
	if err != nil {
		http.Error(w, "bad guild", http.StatusBadRequest)
		return 0, false
	}
	perms, err := b.Discord.Rest.GetGuildMemberPermissions(guildID, snowflake.ID(session.UserID))
	if err != nil || !b.canManage(snowflake.ID(session.UserID), guildID, perms, false) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return 0, false
	}
	return guildID, true
}

func (b *Bot) canManage(userID snowflake.ID, _ snowflake.ID, perms discord.Permissions, owner bool) bool {
	if slices.Contains(b.Config.Bot.DevIDs, userID) || owner || perms.Has(discord.PermissionAdministrator) || perms.Has(discord.PermissionManageGuild) {
		return true
	}
	return false
}

func (b *Bot) authorizedSession(r *http.Request) (dbsqlc.Session, bool) {
	userID, ok := b.authorizedUser(r)
	if !ok {
		return dbsqlc.Session{}, false
	}
	session, err := b.Database.GetSession(r.Context(), int64(userID))
	if err != nil {
		return dbsqlc.Session{}, false
	}
	return session, true
}

func (b *Bot) authorizedUser(r *http.Request) (snowflake.ID, bool) {
	token := strings.TrimSpace(r.Header.Get("Authorization"))
	token = strings.TrimPrefix(token, "Bearer ")
	userID, ok := parseJWT(token, b.Config.Dashboard.SigningKey)
	return userID, ok
}

func tagJSON(tag dbsqlc.GuildTag) map[string]any {
	return map[string]any{
		"id":         tag.ID,
		"name":       tag.Name,
		"user_id":    snowflake.ID(tag.UserID).String(),
		"content":    tag.Content,
		"command_id": strconv.FormatInt(tag.CommandID, 10),
	}
}

func idString(id int64) string {
	return strconv.FormatInt(id, 10)
}

func settingValue(key string, value any) (any, bool) {
	switch key {
	case "announcement_channel_id", "dj_role_id":
		switch typed := value.(type) {
		case string:
			parsed, err := strconv.ParseInt(typed, 10, 64)
			return parsed, err == nil
		case float64:
			return int64(typed), true
		}
	case "join_message", "leave_message":
		text, ok := value.(string)
		return text, ok
	case "join_messages_enabled", "leave_messages_enabled", "nsfw_enabled", "role_saver_enabled":
		enabled, ok := value.(bool)
		return enabled, ok
	}
	return nil, false
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func signJWT(userID snowflake.ID, key string) string {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	payload := base64.RawURLEncoding.EncodeToString([]byte(fmt.Sprintf(`{"sub":"%s","iat":%d}`, userID, time.Now().Unix())))
	return header + "." + payload + "." + signHS256(header+"."+payload, key)
}

func parseJWT(token string, key string) (snowflake.ID, bool) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 || signHS256(parts[0]+"."+parts[1], key) != parts[2] {
		return 0, false
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return 0, false
	}
	var body struct {
		Subject string `json:"sub"`
	}
	if err := json.Unmarshal(payload, &body); err != nil {
		return 0, false
	}
	id, err := snowflake.Parse(body.Subject)
	return id, err == nil
}

func signHS256(message string, key string) string {
	mac := hmac.New(sha256.New, []byte(key))
	_, _ = mac.Write([]byte(message))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func randomToken() string {
	var buf [16]byte
	_, _ = rand.Read(buf[:])
	return hex.EncodeToString(buf[:])
}
