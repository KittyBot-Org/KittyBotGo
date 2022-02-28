package types

import (
	"github.com/DisgoOrg/disgolink/lavalink"
	"github.com/DisgoOrg/snowflake"
)

var _ lavalink.PlayerEventListener = (*MusicPlayer)(nil)

type MusicPlayer struct {
	lavalink.Player
	Bot       *Bot
	Type      PlayerType
	Queue     *MusicQueue
	History   *MusicHistory
	SkipVotes map[snowflake.Snowflake]struct{}
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

		case LoopingTypeRepeatSong:
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

}

func (p *MusicPlayer) OnTrackStuck(player lavalink.Player, track lavalink.AudioTrack, thresholdMs lavalink.Duration) {

}

func (p *MusicPlayer) OnWebSocketClosed(player lavalink.Player, code int, reason string, byRemote bool) {

}

type PlayerType int

const (
	PlayerTypeMusic PlayerType = iota
	PlayerTypeRadio
)
