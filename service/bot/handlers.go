package bot

import (
	"context"
	"log/slog"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgolink/v4/disgolink"
	"github.com/disgoorg/lavaqueue-plugin"
	"github.com/topi314/tint"
)

func (b *Bot) OnDiscordEvent(event bot.Event) {
	switch e := event.(type) {
	case *events.VoiceServerUpdate:
		slog.Debug("received voice server update")
		if e.Endpoint == nil {
			return
		}
		b.Lavalink.OnVoiceServerUpdate(context.Background(), e.GuildID, e.Token, *e.Endpoint)
	case *events.GuildVoiceStateUpdate:
		if e.VoiceState.UserID != b.Discord.ApplicationID {
			return
		}
		slog.Debug("received voice state update")
		b.Lavalink.OnVoiceStateUpdate(context.Background(), e.VoiceState.GuildID, e.VoiceState.ChannelID, e.VoiceState.SessionID)
	case *events.GuildReady:
		go b.cacheGuildInvites(e.GuildID)
	case *events.GuildJoin:
		go b.cacheGuildInvites(e.GuildID)
	case *events.GuildLeave:
		b.inviteMu.Lock()
		delete(b.invites, e.GuildID)
		b.inviteMu.Unlock()
		b.usedMu.Lock()
		delete(b.used, e.GuildID)
		b.usedMu.Unlock()
	case *events.InviteCreate:
		if e.GuildID != nil {
			b.rememberInvite(*e.GuildID, e.Code, e.Uses)
		}
	case *events.InviteDelete:
		if e.GuildID != nil {
			b.forgetInvite(*e.GuildID, e.Code)
		}
	case *events.GuildMemberJoin:
		b.onMemberJoin(e)
	case *events.GuildMemberLeave:
		b.onMemberLeave(e)
	case *events.GuildMemberUpdate:
		b.onMemberUpdate(e)
	case *events.InteractionCreate:
		b.onPublishedTag(e)
	}
}

func (b *Bot) OnLavalinkEvent(event disgolink.Event) {
	switch e := event.(type) {
	case lavaqueue.QueueEndEvent:
		slog.Info("queue end", slog.String("guild", e.GuildID().String()))

	case *disgolink.PlayerTrackExceptionEvent:
		slog.Error("track exception", tint.Err(e.Exception))

	case *disgolink.PlayerTrackStartEvent, *disgolink.PlayerTrackStuckEvent:
	}
}
