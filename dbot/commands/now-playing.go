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
		ComponentHandlers: map[string]handler.ComponentHandler{
			"previous":   previousComponentHandler(b),
			"play-pause": playPauseComponentHandler,
			"next":       nextComponentHandler,
			"like":       likeComponentHandler,
		},
	}
}

func nowPlayingHandler(b *dbot.Bot) handler.CommandHandler {
	return func(e *events.ApplicationCommandInteractionCreate) error {
		player := b.MusicPlayers.Get(*e.GuildID())
		track := player.PlayingTrack()

		if track == nil {
			return e.CreateMessage(responses.CreateErrorf("modules.music.commands.nowplaying.no.track"))
		}
		i := track.Info()
		embed := discord.NewEmbedBuilder().
			SetAuthorName(p.Sprintf("modules.music.commands.nowplaying.title")).
			SetTitle(i.Title).
			SetURL(*i.URI).
			AddField("modules.music.commands.nowplaying.author", i.Author, true).
			AddField("modules.music.commands.nowplaying.requested.by", discord.UserMention(track.UserData().(dbot.AudioTrackData).Requester), true).
			AddField("modules.music.commands.nowplaying.volume", fmt.Sprintf("%d%%", player.Volume()), true).
			SetThumbnail(getArtworkURL(player.PlayingTrack())).
			SetFooterText(fmt.Sprintf("modules.music.commands.nowplaying.footer", player.Queue.Len()))
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
