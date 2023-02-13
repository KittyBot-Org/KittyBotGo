package bot

import (
	"github.com/disgoorg/disgolink/v2/disgolink"
	"github.com/disgoorg/disgolink/v2/lavalink"
	"github.com/disgoorg/json"
)

type Player struct {
	disgolink.Player `json:"-"`

	NodeName string `json:"node_name"`
	Queue    *Queue `json:"queue"`
}

func (p Player) MarshalJSON() ([]byte, error) {
	p.NodeName = p.Node().Config().Name

	type player Player
	return json.Marshal(player(p))
}

type QueueType int

const (
	QueueTypeDefault QueueType = iota
	QueueTypeLoopTrack
	QueueTypeLoopQueue
)

type Queue struct {
	Type   QueueType `json:"type"`
	Tracks []lavalink.Track
}

func (q *Queue) Next() (lavalink.Track, bool) {
	if len(q.Tracks) == 0 {
		return lavalink.Track{}, false
	}

	track := q.Tracks[0]
	q.Tracks = q.Tracks[1:]

	return track, true
}

func (q *Queue) Add(track []lavalink.Track) {
	q.Tracks = append(q.Tracks, track...)
}

func (q *Queue) Remove(i int) {
	q.Tracks = append(q.Tracks[:i], q.Tracks[i+1:]...)
}

func (q *Queue) Clear() {
	q.Tracks = nil
}

func (q *Queue) For(f func(track lavalink.Track)) {
	for _, track := range q.Tracks {
		f(track)
	}
}
