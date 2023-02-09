package commands

import (
	"context"
	"regexp"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/disgolink/v2/disgolink"
	"github.com/disgoorg/disgolink/v2/lavalink"
	"github.com/disgoorg/json"

	"github.com/KittyBot-Org/KittyBotGo/service/bot"
)

var (
	urlPattern    = regexp.MustCompile("^https?://[-a-zA-Z0-9+&@#/%?=~_|!:,.;]*[-a-zA-Z0-9+&@#/%=~_|]?")
	searchPattern = regexp.MustCompile(`^(.{2})(search|isrc):(.+)`)
)

var play = discord.SlashCommandCreate{
	Name:        "play",
	Description: "Play a song",
	Options: []discord.ApplicationCommandOption{
		discord.ApplicationCommandOptionString{
			Name:        "query",
			Description: "The song or search to play",
			Required:    true,
		},
		discord.ApplicationCommandOptionString{
			Name:        "source",
			Description: "The source to search on",
			Choices: []discord.ApplicationCommandOptionChoiceString{
				{
					Name:  "YouTube",
					Value: string(lavalink.SearchTypeYouTube),
				},
				{
					Name:  "YouTube Music",
					Value: string(lavalink.SearchTypeYouTubeMusic),
				},
				{
					Name:  "SoundCloud",
					Value: string(lavalink.SearchTypeSoundCloud),
				},
				{
					Name:  "Deezer",
					Value: "dzsearch",
				},
				{
					Name:  "Deezer ISRC",
					Value: "dzisrc",
				},
			},
		},
	},
}

func (b *Bot) OnPlay(e *handler.CommandEvent) error {
	data := e.SlashCommandInteractionData()

	query := data.String("query")

	if source, ok := data.OptString("source"); ok {
		query = lavalink.SearchType(source).Apply(query)
	} else {
		if !urlPattern.MatchString(query) && !searchPattern.MatchString(query) {
			query = lavalink.SearchTypeYouTube.Apply(query)
		}
	}

	voiceState, ok := b.Discord.Caches().VoiceState(*e.GuildID(), e.User().ID)
	if !ok {
		return e.CreateMessage(discord.MessageCreate{
			Content: "You must be in a voice channel",
			Flags:   discord.MessageFlagEphemeral,
		})
	}

	_, ok = b.Discord.Caches().VoiceState(*e.GuildID(), e.ApplicationID())
	if !ok {
		ctx := bot.ContextWithShardID(context.Background(), e.ShardID())
		_ = b.Discord.UpdateVoiceState(ctx, *e.GuildID(), voiceState.ChannelID, false, false)
	}

	_ = e.DeferCreateMessage(false)

	player := b.Lavalink.Player(*e.GuildID())

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	player.Node().LoadTracks(ctx, query, disgolink.NewResultHandler(
		func(track lavalink.Track) {
			_ = player.Update(ctx, lavalink.WithTrack(track))
			_, _ = e.UpdateInteractionResponse(discord.MessageUpdate{
				Content: json.Ptr("Playing: " + track.Info.Title),
			})
		},
		func(playlist lavalink.Playlist) {
			_ = player.Update(ctx, lavalink.WithTrack(playlist.Tracks[0]))
			_, _ = e.UpdateInteractionResponse(discord.MessageUpdate{
				Content: json.Ptr("Playing: " + playlist.Tracks[0].Info.Title),
			})
		},
		func(tracks []lavalink.Track) {
			_ = player.Update(ctx, lavalink.WithTrack(tracks[0]))
			_, _ = e.UpdateInteractionResponse(discord.MessageUpdate{
				Content: json.Ptr("Playing: " + tracks[0].Info.Title),
			})
		},
		func() {
			_, _ = e.UpdateInteractionResponse(discord.MessageUpdate{
				Content: json.Ptr("No tracks found"),
			})
		},
		func(err error) {
			_, _ = e.UpdateInteractionResponse(discord.MessageUpdate{
				Content: json.Ptr("An error occurred: " + err.Error()),
			})
		},
	))

	return nil
}
