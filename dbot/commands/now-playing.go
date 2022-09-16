package commands

import (
	"fmt"

	"github.com/KittyBot-Org/KittyBotGo/dbot"
	"github.com/KittyBot-Org/KittyBotGo/dbot/responses"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/handler"
)

func NowPlaying(b *dbot.Bot) handler.Command {
	return handler.Command{
		Create: discord.SlashCommandCreate{
			Name:        "now-playing",
			Description: "Tells you about the currently playing song.",
		},
		Check: dbot.HasMusicPlayer(b).And(dbot.IsPlaying(b)),
		CommandHandlers: map[string]handler.CommandHandler{
			"": nowPlayingHandler(b),
		},
	}
}

func nowPlayingHandler(b *dbot.Bot) handler.CommandHandler {
	return func(e *events.ApplicationCommandInteractionCreate) error {
		player := b.MusicPlayers.Get(*e.GuildID())
		track := player.PlayingTrack()

		if track == nil {
			return e.CreateMessage(responses.CreateErrorf("There is no song playing right now."))
		}
		i := track.Info()
		embed := discord.NewEmbedBuilder().
			SetAuthorName("Currently playing:").
			SetTitle(i.Title).
			SetURL(*i.URI).
			AddField("Author:", i.Author, true).
			AddField("Requested by:", discord.UserMention(track.UserData().(dbot.AudioTrackData).Requester), true).
			AddField("Volume:", fmt.Sprintf("%d%%", player.Volume()), true).
			SetThumbnail(getArtworkURL(player.PlayingTrack())).
			SetFooterText(fmt.Sprintf("Tracks in queue: %d", player.Queue.Len()))
		if !i.IsStream {
			bar := [10]string{"▬", "▬", "▬", "▬", "▬", "▬", "▬", "▬", "▬", "▬"}
			t1 := player.Position()
			t2 := i.Length
			p := int(float64(t1) / float64(t2) * 10)
			bar[p] = "🔘"
			loopString := ""
			if player.Queue.LoopingType() == dbot.LoopingTypeRepeatSong {
				loopString = "🔂"
			}
			if player.Queue.LoopingType() == dbot.LoopingTypeRepeatQueue {
				loopString = "🔁"
			}
			embed.Description += fmt.Sprintf("\n\n%s / %s %s\n%s", formatPosition(t1), formatPosition(t2), loopString, bar)
		}
		return e.CreateMessage(responses.CreateSuccessEmbedComponents(embed.Build(), getMusicControllerComponents(track)))
	}
}
