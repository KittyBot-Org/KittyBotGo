package commands

import (
	"encoding/json"

	"github.com/disgoorg/disgolink/v4/lavalink"
)

func decodeTrack(data []byte) (lavalink.Track, error) {
	var track lavalink.Track
	err := json.Unmarshal(data, &track)
	return track, err
}
