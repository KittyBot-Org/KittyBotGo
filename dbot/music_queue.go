package dbot

import (
	"math/rand"
	"sync"

	"github.com/disgoorg/disgolink/lavalink"
)

type LoopingType int

const (
	LoopingTypeOff LoopingType = iota
	LoopingTypeRepeatSong
	LoopingTypeRepeatQueue
)

func (t LoopingType) String() string {
	switch t {
	case LoopingTypeOff:
		return "Off"
	case LoopingTypeRepeatSong:
		return "Repeat Song"
	case LoopingTypeRepeatQueue:
		return "Repeat Queue"
	default:
		return "unknown"
	}
}

func NewMusicQueue(queueType LoopingType) *MusicQueue {
	return &MusicQueue{
		loopingType: queueType,
	}
}

type MusicQueue struct {
	mu          sync.Mutex
	tracks      []lavalink.AudioTrack
	loopingType LoopingType
}

func (q *MusicQueue) LoopingType() LoopingType {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.loopingType
}

func (q *MusicQueue) SetType(loopingType LoopingType) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.loopingType = loopingType
}

func (q *MusicQueue) Shuffle() {
	q.mu.Lock()
	defer q.mu.Unlock()
	for i := range q.tracks {
		j := rand.Intn(i + 1)
		q.tracks[i], q.tracks[j] = q.tracks[j], q.tracks[i]
	}
}

func (q *MusicQueue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.tracks)
}

func (q *MusicQueue) Tracks() []lavalink.AudioTrack {
	q.mu.Lock()
	defer q.mu.Unlock()
	items := make([]lavalink.AudioTrack, len(q.tracks))
	for i := range q.tracks {
		items[i] = q.tracks[i]
	}
	return items
}

func (q *MusicQueue) Clear() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.tracks = nil
}

func (q *MusicQueue) Pop() lavalink.AudioTrack {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.tracks) == 0 {
		return nil
	}
	var item lavalink.AudioTrack
	item, q.tracks = q.tracks[0], q.tracks[1:]
	return item
}

func (q *MusicQueue) Push(tracks ...lavalink.AudioTrack) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.tracks = append(q.tracks, tracks...)
}

func (q *MusicQueue) Get(index int) lavalink.AudioTrack {
	q.mu.Lock()
	defer q.mu.Unlock()
	if index < 0 || index >= len(q.tracks) {
		return nil
	}
	return q.tracks[index]
}

func (q *MusicQueue) Remove(index ...int) {
	q.mu.Lock()
	defer q.mu.Unlock()
	for _, i := range index {
		if i < len(q.tracks) {
			q.tracks = append(q.tracks[:i], q.tracks[i+1:]...)
		}
	}
}
