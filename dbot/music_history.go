package dbot

import (
	"sync"

	"github.com/disgoorg/disgolink/lavalink"
)

func NewHistory() *MusicHistory {
	return &MusicHistory{}
}

type MusicHistory struct {
	mu     sync.Mutex
	tracks []lavalink.AudioTrack
}

func (h *MusicHistory) Len() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.tracks)
}

func (h *MusicHistory) Tracks() []lavalink.AudioTrack {
	h.mu.Lock()
	defer h.mu.Unlock()
	items := make([]lavalink.AudioTrack, len(h.tracks))
	for i := range h.tracks {
		items[i] = h.tracks[i]
	}
	return items
}

func (h *MusicHistory) Push(tracks ...lavalink.AudioTrack) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.tracks = append(h.tracks, tracks...)
}

func (h *MusicHistory) Last() lavalink.AudioTrack {
	h.mu.Lock()
	defer h.mu.Unlock()
	if len(h.tracks) == 0 {
		return nil
	}
	return h.tracks[len(h.tracks)-1]
}
