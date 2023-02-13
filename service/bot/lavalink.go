package bot

import (
	"context"

	"github.com/disgoorg/disgolink/v2/disgolink"
	"github.com/disgoorg/disgolink/v2/lavalink"
	"github.com/disgoorg/snowflake/v2"

	"github.com/KittyBot-Org/KittyBotGo/interal/config"
)

func (b *Bot) OnLavalinkEvent(p disgolink.Player, event lavalink.Event) {
	player := b.Player(p.GuildID())
	switch e := event.(type) {
	case lavalink.TrackStartEvent:

	case lavalink.TrackEndEvent:
		if !e.Reason.MayStartNext() {
			return
		}
		track, ok := player.Queue.Next()
		if !ok {
			if err := player.Destroy(context.Background()); err != nil {
				b.Logger.Error("failed to destroy player: ", err)
			}
			return
		}
		if err := player.Update(context.Background(), lavalink.WithTrack(track)); err != nil {
			b.Logger.Error("failed to update player: ", err)
		}

	case lavalink.TrackExceptionEvent:
		b.Logger.Debug("received track exception event")

	case lavalink.TrackStuckEvent:
		b.Logger.Debug("received track stuck event")

	}
}

func (b *Bot) SavePlayers() {
	if err := config.Save("players.json", b.Players); err != nil {
		b.Logger.Error("failed to save players", err)
	}
}

func (b *Bot) LoadPlayers() {
	var players map[snowflake.ID]*Player
	if err := config.Load("players.json", &players); err != nil {
		b.Logger.Error("failed to load players: ", err)
	}
	for guildID, player := range players {
		player.Player = b.Lavalink.PlayerOnNode(player.NodeName, guildID)
		if err := player.Sync(context.Background()); err != nil {
			b.Logger.Error("failed to sync player: ", err)
			continue
		}
		voiceState, ok := b.Discord.Caches().VoiceState(guildID, b.Discord.ApplicationID())
		if !ok {
			b.Logger.Error("failed to get voice state")
			continue
		}
		player.OnVoiceStateUpdate(context.Background(), voiceState.ChannelID, voiceState.SessionID)
		b.Players[guildID] = player
	}
}
