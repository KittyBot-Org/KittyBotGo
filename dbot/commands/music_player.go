package commands

import (
	"github.com/KittyBot-Org/KittyBotGo/dbot"
	"github.com/KittyBot-Org/KittyBotGo/dbot/responses"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/handler"
	"github.com/go-jet/jet/v2/qrm"
)

func PlayerNext(b *dbot.Bot) handler.Component {
	return handler.Component{
		Name:    "next",
		Check:   nil,
		Handler: nextComponentHandler(b),
	}
}

func nextComponentHandler(b *dbot.Bot) handler.ComponentHandler {
	return func(e *events.ComponentInteractionCreate) error {
		player, err := checkPlayer(b, e)
		if player == nil {
			return err
		}
		nextTrack := player.Queue.Pop()
		if nextTrack == nil {
			return e.CreateMessage(responses.CreateErrorf("No songs in the queue."))
		}

		if err = player.Play(nextTrack); err != nil {
			return e.CreateMessage(responses.CreateErrorf("Failed to play next song. Please try again."))
		}
		return e.UpdateMessage(responses.UpdateSuccessComponentsf("Skipped to next song.", []any{formatTrack(nextTrack), nextTrack.Info().Length}, getMusicControllerComponents(nextTrack)))
	}
}

func PlayerPlayPause(b *dbot.Bot) handler.Component {
	return handler.Component{
		Name:    "play-pause",
		Check:   nil,
		Handler: playPauseComponentHandler(b),
	}
}

func playPauseComponentHandler(b *dbot.Bot) handler.ComponentHandler {
	return func(e *events.ComponentInteractionCreate) error {
		player, err := checkPlayer(b, e)
		if player == nil {
			return err
		}
		if player.PlayingTrack() == nil {
			return e.CreateMessage(responses.CreateErrorf("There is currently no track playing."))
		}
		paused := !player.Paused()
		if err = player.Pause(paused); err != nil {
			if paused {
				return e.CreateMessage(responses.CreateErrorf("Failed to pause the music player. Please try again."))
			}
			return e.CreateMessage(responses.CreateErrorf("Failed to play the music player. Please try again."))
		}
		track := player.PlayingTrack()
		if paused {
			return e.UpdateMessage(responses.UpdateSuccessComponentsf("⏸ Paused the music player.\nPaused: %s - %s at `%s`.", []any{formatTrack(track), track.Info().Length, player.Position()}, getMusicControllerComponents(track)))
		}
		return e.UpdateMessage(responses.UpdateSuccessComponentsf("▶ Resumed the music player.\nPlaying: %s - %s", []any{formatTrack(track), track.Info().Length}, getMusicControllerComponents(track)))
	}
}

func PlayerPrevious(b *dbot.Bot) handler.Component {
	return handler.Component{
		Name:    "previous",
		Handler: previousComponentHandler(b),
	}
}

func previousComponentHandler(b *dbot.Bot) handler.ComponentHandler {
	return func(e *events.ComponentInteractionCreate) error {
		player, err := checkPlayer(b, e)
		if player == nil {
			return err
		}
		nextTrack := player.History.Last()
		if nextTrack == nil {
			return e.CreateMessage(responses.CreateErrorf("No songs in the history."))
		}

		if err = player.Play(nextTrack); err != nil {
			return e.CreateMessage(responses.CreateErrorf("Failed to play previous song. Please try again."))
		}
		return e.UpdateMessage(responses.UpdateSuccessComponentsf("Went back to previous song.", []any{formatTrack(nextTrack), nextTrack.Info().Length}, getMusicControllerComponents(nextTrack)))
	}
}

func PlayerLike(b *dbot.Bot) handler.Component {
	return handler.Component{
		Name:    "like",
		Handler: likeComponentHandler(b),
	}
}

func likeComponentHandler(b *dbot.Bot) handler.ComponentHandler {
	return func(e *events.ComponentInteractionCreate) error {
		if len(e.Message.Embeds) == 0 {
			return e.CreateMessage(responses.CreateErrorf("No music embed found in this message."))
		}
		allMatches := trackRegex.FindAllStringSubmatch(e.Message.Embeds[0].Description, -1)
		if allMatches == nil {
			return e.CreateMessage(responses.CreateErrorf("No track found to like in this message."))
		}
		matches := allMatches[0]
		var (
			title string
			url   string
		)
		title = matches[trackRegex.SubexpIndex("title")]
		if len(matches) > 2 {
			url = matches[trackRegex.SubexpIndex("url")]
		}

		_, err := b.DB.LikedSongs().Get(e.User().ID, title)
		if err != nil && err != qrm.ErrNoRows {
			return err
		}

		if err == qrm.ErrNoRows {
			if err = b.DB.LikedSongs().Add(e.User().ID, getTrackQuery(title, url), title); err != nil {
				b.Logger.Error("Error adding music history entry: ", err)
				return e.CreateMessage(responses.CreateErrorf("Failed to add song to liked songs. Please try again."))
			}
			res := responses.CreateSuccessf("👍 Added [`%s`](%s) to your liked songs.", title, url)
			res.Flags = discord.MessageFlagEphemeral
			return e.CreateMessage(res)

		}
		if err = b.DB.LikedSongs().Delete(e.User().ID, title); err != nil {
			b.Logger.Error("Error removing music history entry: ", err)
			return e.CreateMessage(responses.CreateErrorf("Failed to remove song from your liked songs. Please try again."))
		}
		res := responses.CreateSuccessf("👎 Removed [`%s`](%s) from your liked songs.", title, url)
		res.Flags = discord.MessageFlagEphemeral
		return e.CreateMessage(res)
	}
}
