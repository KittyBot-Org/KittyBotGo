package bot

import (
	"context"
	"log/slog"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgolink/v3/disgolink"
	"github.com/disgoorg/disgolink/v3/lavalink"
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
		if e.VoiceState.UserID != b.Discord.ApplicationID() {
			return
		}
		slog.Debug("received voice state update")
		b.Lavalink.OnVoiceStateUpdate(context.Background(), e.VoiceState.GuildID, e.VoiceState.ChannelID, e.VoiceState.SessionID)
	}
}

func (b *Bot) OnLavalinkEvent(p disgolink.Player, event lavalink.Event) {
	// player := b.Lavalink.Player(p.GuildID())
	switch e := event.(type) {
	case lavaqueue.QueueEndEvent:
		slog.Info("queue end", slog.String("guild", p.GuildID().String()))

	case lavalink.TrackStartEvent:

	case lavalink.TrackExceptionEvent:
		slog.Error("track exception", tint.Err(e.Exception))

	case lavalink.TrackStuckEvent:

	}
}
