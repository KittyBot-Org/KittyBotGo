package commands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/handler"
	"github.com/lithammer/fuzzysearch/fuzzy"

	"github.com/KittyBot-Org/KittyBotGo/dbot"
)

func Remove(b *dbot.Bot) handler.Command {
	return handler.Command{
		Create: discord.SlashCommandCreate{
			Name:        "remove",
			Description: "Removes tracks from the queue.",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionSubCommand{
					Name:        "track",
					Description: "Removes a tracks from the queue.",
					Options: []discord.ApplicationCommandOption{
						discord.ApplicationCommandOptionString{
							Name:         "track",
							Description:  "The track to remove",
							Required:     true,
							Autocomplete: true,
						},
					},
				},
				discord.ApplicationCommandOptionSubCommand{
					Name:        "user-tracks",
					Description: "Removes all tracks from a user from the queue.",
					Options: []discord.ApplicationCommandOption{
						discord.ApplicationCommandOptionUser{
							Name:        "user",
							Description: "From which user to remove the tracks",
							Required:    true,
						},
					},
				},
			},
		},
		Check: dbot.HasMusicPlayer(b).And(dbot.IsMemberConnectedToVoiceChannel(b)).And(dbot.HasQueueItems(b)),
		CommandHandlers: map[string]handler.CommandHandler{
			"track":       removeTrackHandler(b),
			"user-tracks": removeUserTracksHandler(b),
		},
		AutocompleteHandlers: map[string]handler.AutocompleteHandler{
			"track": removeTrackAutocompleteHandler(b),
		},
	}
}

func removeTrackHandler(b *dbot.Bot) handler.CommandHandler {
	return func(e *events.ApplicationCommandInteractionCreate) error {
		player := b.MusicPlayers.Get(*e.GuildID())
		strIndex := e.SlashCommandInteractionData().String("track")
		index, err := strconv.Atoi(strIndex)
		if err != nil {
			return e.CreateMessage(discord.MessageCreate{
				Content: fmt.Sprintf("Invalid track index: `%d`.", index),
				Flags:   discord.MessageFlagEphemeral,
			})
		}

		removeTrack := player.Queue.Get(index - 1)
		if removeTrack == nil {
			return e.CreateMessage(discord.MessageCreate{
				Content: fmt.Sprintf("No track found with index `%d`.", index),
				Flags:   discord.MessageFlagEphemeral,
			})
		}

		player.Queue.Remove(index - 1)
		return e.CreateMessage(discord.MessageCreate{
			Content: fmt.Sprintf("Removed track %s at index `%d` from the queue.", formatTrack(removeTrack), index),
		})
	}
}

func removeUserTracksHandler(b *dbot.Bot) handler.CommandHandler {
	return func(e *events.ApplicationCommandInteractionCreate) error {
		player := b.MusicPlayers.Get(*e.GuildID())
		userID := e.SlashCommandInteractionData().Snowflake("user")

		removedTracks := 0
		for i, track := range player.Queue.Tracks() {
			if track.UserData().(dbot.AudioTrackData).Requester == userID {
				player.Queue.Remove(i - removedTracks)
				removedTracks++
			}
		}
		var msg string
		if removedTracks == 0 {
			msg = fmt.Sprintf("No track from %s found.", discord.UserMention(userID))
		} else {
			msg = fmt.Sprintf("Removed `%d` tracks from %s.", removedTracks, discord.UserMention(userID))
		}

		return e.CreateMessage(discord.NewMessageCreateBuilder().
			SetContent(msg).
			SetAllowedMentions(&discord.AllowedMentions{}).
			Build(),
		)
	}
}

func removeTrackAutocompleteHandler(b *dbot.Bot) handler.AutocompleteHandler {
	return func(e *events.AutocompleteInteractionCreate) error {
		player := b.MusicPlayers.Get(*e.GuildID())
		if player == nil || player.Queue.Len() == 0 {
			return e.Result(nil)
		}
		tracks := make([]string, player.Queue.Len())

		for i, track := range player.Queue.Tracks() {
			tracks[i] = fmt.Sprintf("%d. %s", i+1, track.Info().Title)
		}

		ranks := fuzzy.RankFindFold(e.Data.String("track"), tracks)

		choicesLen := len(ranks)
		if choicesLen > 25 {
			choicesLen = 25
		}
		choices := make([]discord.AutocompleteChoice, choicesLen)

		for i, rank := range ranks {
			if i >= 25 {
				break
			}
			choices[i] = discord.AutocompleteChoiceString{
				Name:  rank.Target,
				Value: strings.SplitN(rank.Target, ".", 2)[0],
			}
		}
		return e.Result(choices)
	}
}
