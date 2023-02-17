package res

import (
	"fmt"

	"github.com/disgoorg/disgolink/v2/lavalink"
)

func FormatTrack(track lavalink.Track, position lavalink.Duration) string {
	var positionStr string
	if position > 0 {
		positionStr = fmt.Sprintf("`%s/%s`", FormatDuration(position), FormatDuration(track.Info.Length))
	} else {
		positionStr = fmt.Sprintf("`%s`", FormatDuration(track.Info.Length))
	}

	if track.Info.URI != nil {
		return fmt.Sprintf("[`%s`](<%s>) - `%s` %s", track.Info.Title, *track.Info.URI, track.Info.Author, positionStr)
	}
	return fmt.Sprintf("`%s` - `%s` %s`", track.Info.Title, track.Info.Author, positionStr)
}

func FormatDuration(duration lavalink.Duration) string {
	if duration == 0 {
		return "00:00"
	}
	return fmt.Sprintf("%02d:%02d", duration.Minutes(), duration.SecondsPart())
}

func Trim(s string, length int) string {
	r := []rune(s)
	if len(r) > length {
		return string(r[:length-1]) + "…"
	}
	return s
}
