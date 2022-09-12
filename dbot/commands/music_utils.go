package commands

import (
	"fmt"

	"github.com/KittyBot-Org/KittyBotGo/dbot"
	"github.com/KittyBot-Org/KittyBotGo/dbot/responses"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgolink/lavalink"
	source_plugins "github.com/disgoorg/source-plugins"
	"github.com/go-jet/jet/v2/qrm"
)

func getMusicControllerComponents(track lavalink.AudioTrack) discord.ContainerComponent {
	buttons := discord.ActionRowComponent{
		discord.NewPrimaryButton("", "cmd:now-playing:previous").WithEmoji(discord.ComponentEmoji{Name: "⏮"}),
		discord.NewPrimaryButton("", "cmd:now-playing:play-pause").WithEmoji(discord.ComponentEmoji{Name: "⏯"}),
		discord.NewPrimaryButton("", "cmd:now-playing:next").WithEmoji(discord.ComponentEmoji{Name: "⏭"}),
	}
	if track != nil {
		buttons = buttons.AddComponents(discord.NewPrimaryButton("", "cmd:now-playing:like").WithEmoji(discord.ComponentEmoji{Name: "❤"}))
	}
	return buttons
}

func formatTrack(track lavalink.AudioTrack) string {
	return fmt.Sprintf("[`%s`](%s)", track.Info().Title, *track.Info().URI)
}

func getTrackQuery(title string, url *string) string {
	if url != nil {
		return *url
	}
	return title
}

func checkPlayer(b *dbot.Bot, e *events.ComponentInteractionCreate) (*dbot.MusicPlayer, error) {
	player := b.MusicPlayers.Get(*e.GuildID())
	if player == nil {
		return nil, e.CreateMessage(responses.CreateErrorf("modules.music.components.no.player"))
	}
	return player, nil
}

func formatPosition(position lavalink.Duration) string {
	if position == 0 {
		return "0:00"
	}
	return fmt.Sprintf("%d:%02d", position.Minutes(), position.SecondsPart())
}

func getArtworkURL(track lavalink.AudioTrack) string {
	switch track.Info().SourceName {
	case "youtube":
		return "https://i.ytimg.com/vi/" + track.Info().Identifier + "/maxresdefault.jpg"

	case "twitch":
		return "https://static-cdn.jtvnw.net/previews-ttv/live_user_" + track.Info().Identifier + "-440x248.jpg"

	case "spotify":
		if spotifyTrack, ok := track.(*source_plugins.SpotifyAudioTrack); ok && spotifyTrack.ArtworkURL != nil {
			return *spotifyTrack.ArtworkURL
		}

	case "applemusic":
		if appleMusicTrack, ok := track.(*source_plugins.AppleMusicAudioTrack); ok && appleMusicTrack.ArtworkURL != nil {
			return *appleMusicTrack.ArtworkURL
		}
	}
	return ""
}

func playPauseComponentHandler(b *dbot.Bot, _ []string, e *events.ComponentInteractionCreate) error {
	player, err := checkPlayer(b, e)
	if player == nil {
		return err
	}
	if player.PlayingTrack() == nil {
		return e.CreateMessage(responses.CreateErrorf("modules.music.components.play.pause.not.playing"))
	}
	paused := !player.Paused()
	if err = player.Pause(paused); err != nil {
		if paused {
			return e.CreateMessage(responses.CreateErrorf("modules.music.components.play.pause.pause.error"))
		}
		return e.CreateMessage(responses.CreateErrorf("modules.music.components.play.pause.play.error"))
	}
	track := player.PlayingTrack()
	if paused {
		return e.UpdateMessage(responses.UpdateSuccessComponentsf("modules.music.components.play.pause.pause.success", []any{formatTrack(track), track.Info().Length, player.Position()}, getMusicControllerComponents(track)))
	}
	return e.UpdateMessage(responses.UpdateSuccessComponentsf("modules.music.components.play.pause.play.success", []any{formatTrack(track), track.Info().Length}, getMusicControllerComponents(track)))
}

func nextComponentHandler(b *dbot.Bot, _ []string, e *events.ComponentInteractionCreate) error {
	player, err := checkPlayer(b, e)
	if player == nil {
		return err
	}
	nextTrack := player.Queue.Pop()
	if nextTrack == nil {
		return e.CreateMessage(responses.CreateErrorf("modules.music.components.next.empty"))
	}

	if err = player.Play(nextTrack); err != nil {
		return e.CreateMessage(responses.CreateErrorf("modules.music.components.next.error"))
	}
	return e.UpdateMessage(responses.UpdateSuccessComponentsf("modules.music.commands.next.success", []any{formatTrack(nextTrack), nextTrack.Info().Length}, getMusicControllerComponents(nextTrack)))
}

func likeComponentHandler(b *dbot.Bot, _ []string, e *events.ComponentInteractionCreate) error {
	if len(e.Message.Embeds) == 0 {
		return e.CreateMessage(responses.CreateErrorf("modules.music.components.like.no.embed"))
	}
	allMatches := trackRegex.FindAllStringSubmatch(e.Message.Embeds[0].Description, -1)
	if allMatches == nil {
		return e.CreateMessage(responses.CreateErrorf("modules.music.components.like.no.track"))
	}
	matches := allMatches[0]
	var (
		title string
		url   *string
	)
	title = matches[trackRegex.SubexpIndex("title")]
	if len(matches) > 2 {
		url = &matches[trackRegex.SubexpIndex("url")]
	}

	_, err := b.DB.LikedSongs().Get(e.User().ID, title)
	if err != nil && err != qrm.ErrNoRows {
		return err
	}

	if err == qrm.ErrNoRows {
		if err = b.DB.LikedSongs().Add(e.User().ID, getTrackQuery(title, url), title); err != nil {
			b.Logger.Error("Error adding music history entry: ", err)
			return e.CreateMessage(responses.CreateErrorf("modules.music.components.like.add.error"))
		}
		res := responses.CreateSuccessf("modules.music.components.like.add.success", title, url)
		res.Flags = discord.MessageFlagEphemeral
		return e.CreateMessage(res)

	}
	if err = b.DB.LikedSongs().Delete(e.User().ID, title); err != nil {
		b.Logger.Error("Error removing music history entry: ", err)
		return e.CreateMessage(responses.CreateErrorf("modules.music.components.like.remove.error"))
	}
	res := responses.CreateSuccessf("modules.music.components.like.remove.success", title, url)
	res.Flags = discord.MessageFlagEphemeral
	return e.CreateMessage(res)
}
