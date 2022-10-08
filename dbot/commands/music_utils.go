package commands

import (
	"fmt"

	"github.com/KittyBot-Org/KittyBotGo/dbot"
	"github.com/KittyBot-Org/KittyBotGo/dbot/responses"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgolink/lavalink"
	source_plugins "github.com/disgoorg/source-plugins"
)

func getMusicControllerComponents(track lavalink.AudioTrack) discord.ContainerComponent {
	buttons := discord.ActionRowComponent{
		discord.NewPrimaryButton("", "now-playing:previous").WithEmoji(discord.ComponentEmoji{Name: "⏮"}),
		discord.NewPrimaryButton("", "now-playing:play-pause").WithEmoji(discord.ComponentEmoji{Name: "⏯"}),
		discord.NewPrimaryButton("", "now-playing:next").WithEmoji(discord.ComponentEmoji{Name: "⏭"}),
	}
	if track != nil {
		buttons = buttons.AddComponents(discord.NewPrimaryButton("", "now-playing:like").WithEmoji(discord.ComponentEmoji{Name: "❤"}))
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
		return nil, e.CreateMessage(responses.CreateErrorf("No music player found in this server."))
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

	case "deezer":
		if deezerTrack, ok := track.(*source_plugins.DeezerAudioTrack); ok && deezerTrack.ArtworkURL != nil {
			return *deezerTrack.ArtworkURL
		}
	}
	return ""
}
