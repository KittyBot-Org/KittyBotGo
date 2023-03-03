package dbot

import (
	"context"
	"time"

	"github.com/disgoorg/disgolink/lavalink"
	"github.com/disgoorg/snowflake/v2"
)

var _ lavalink.PlayerEventListener = (*MusicPlayer)(nil)

type AudioTrackData struct {
	Requester snowflake.ID `json:"requester"`
}

type MusicPlayer struct {
	lavalink.Player
	Bot               *Bot
	Type              PlayerType
	Queue             *MusicQueue
	History           *MusicHistory
	SkipVotes         map[snowflake.ID]struct{}
	DisconnectContext context.Context
	DisconnectCancel  context.CancelFunc
}

func (p *MusicPlayer) OnPlayerPause(player lavalink.Player) {

}

func (p *MusicPlayer) OnPlayerResume(player lavalink.Player) {

}

func (p *MusicPlayer) OnPlayerUpdate(player lavalink.Player, state lavalink.PlayerState) {

}

func (p *MusicPlayer) OnTrackStart(player lavalink.Player, track lavalink.AudioTrack) {

}

func (p *MusicPlayer) OnTrackEnd(player lavalink.Player, track lavalink.AudioTrack, endReason lavalink.AudioTrackEndReason) {
	p.History.Push(track.Clone())
	if endReason.MayStartNext() {
		var nextTrack lavalink.AudioTrack
		switch p.Queue.LoopingType() {
		case LoopingTypeOff:
			if p.Queue.Len() == 0 {
				return
			}
			nextTrack = p.Queue.Pop()

		case LoopingTypeRepeatTrack:
			nextTrack = track.Clone()

		case LoopingTypeRepeatQueue:
			p.Queue.Push(track.Clone())
			nextTrack = p.Queue.Pop()
		}
		if err := player.Play(nextTrack); err != nil {
			p.Bot.Logger.Error("Failed to play next track: ", err)
		}
	}
}

func (p *MusicPlayer) OnTrackException(player lavalink.Player, track lavalink.AudioTrack, exception lavalink.FriendlyException) {
	p.Bot.Logger.Error("Track exception: ", exception.Error())
}

func (p *MusicPlayer) OnTrackStuck(player lavalink.Player, track lavalink.AudioTrack, thresholdMs lavalink.Duration) {

}

func (p *MusicPlayer) OnWebSocketClosed(player lavalink.Player, code int, reason string, byRemote bool) {

}

func (p *MusicPlayer) PlanDisconnect() {
	var ctx context.Context
	ctx, p.DisconnectCancel = context.WithTimeout(context.Background(), 2*time.Minute)
	defer p.DisconnectCancel()

	<-ctx.Done()
	if ctx.Err() == context.DeadlineExceeded {
		if err := p.Bot.Client.UpdateVoiceState(context.TODO(), p.GuildID(), nil, false, false); err != nil {
			p.Bot.Logger.Error("Failed to disconnect from voice channel: ", err)
		}
	}
}

func (p *MusicPlayer) CancelDisconnect() {
	p.DisconnectCancel()
}

type PlayerType int

const (
	PlayerTypeMusic PlayerType = iota
	PlayerTypeRadio
)
