package types

import (
	"sync"

	"github.com/DisgoOrg/disgolink/lavalink"
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

func (h *MusicHistory) All() []lavalink.AudioTrack {
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