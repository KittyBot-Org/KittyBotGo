package bot

import (
	"context"

	"github.com/disgoorg/disgolink/v2/disgolink"
	"github.com/disgoorg/disgolink/v2/lavalink"
)

func (b *Bot) OnLavalinkEvent(p disgolink.Player, event lavalink.Event) {
	player := b.Lavalink.Player(p.GuildID())
	switch e := event.(type) {
	case lavalink.TrackStartEvent:

	case lavalink.TrackEndEvent:
		if err := b.Database.AddHistoryTracks(p.GuildID(), []lavalink.Track{e.Track}); err != nil {
			b.Logger.Error("failed to add history tracks: ", err)
		}
		if !e.Reason.MayStartNext() {
			return
		}

		track, err := b.Database.NextQueueTrack(p.GuildID())
		if err != nil {
			if err = player.Destroy(context.Background()); err != nil {
				b.Logger.Error("failed to destroy player: ", err)
			}
			return
		}
		if err = player.Update(context.Background(), lavalink.WithEncodedTrack(track.Track.Encoded)); err != nil {
			b.Logger.Error("failed to update player: ", err)
		}

	case lavalink.TrackExceptionEvent:
		b.Logger.Debug("received track exception event")

	case lavalink.TrackStuckEvent:
		b.Logger.Debug("received track stuck event")
	}
}

func (b *Bot) RestorePlayers() {
	b.Lavalink.ForPlayers(func(player disgolink.Player) {
		voiceState, ok := b.Discord.Caches().VoiceState(player.GuildID(), b.Discord.ApplicationID())
		if !ok {
			b.Logger.Error("failed to get voice state")
			return
		}
		player.OnVoiceStateUpdate(context.Background(), voiceState.ChannelID, voiceState.SessionID)
	})
}
