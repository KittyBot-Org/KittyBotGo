package commands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/KittyBot-Org/KittyBotGo/dbot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"golang.org/x/text/message"
)

var Remove = handler.Command{
	Create: discord.SlashCommandCreate{
		Name:        "remove",
		Description: "Removes songs from the queue.",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionSubCommand{
				Name:        "song",
				Description: "Removes a songs from the queue.",
				Options: []discord.ApplicationCommandOption{
					discord.ApplicationCommandOptionString{
						Name:         "song",
						Description:  "the song to remove",
						Required:     true,
						Autocomplete: true,
					},
				},
			},
			discord.ApplicationCommandOptionSubCommand{
				Name:        "user-songs",
				Description: "Removes all songs from a user from the queue.",
				Options: []discord.ApplicationCommandOption{
					discord.ApplicationCommandOptionUser{
						Name:        "user",
						Description: "from which user to remove the songs",
						Required:    true,
					},
				},
			},
		},
	},
	Checks: dbot.HasMusicPlayer.And(dbot.IsMemberConnectedToVoiceChannel).And(dbot.HasQueueItems),
	CommandHandler: map[string]handler.CommandHandler{
		"song":       removeSongHandler,
		"user-songs": removeUserSongsHandler,
	},
	AutoCompleteHandler: map[string]dbot.AutocompleteHandler{
		"song": removeSongAutocompleteHandler,
	},
}

func removeSongHandler(b *dbot.Bot, p *message.Printer, e *events.ApplicationCommandInteractionCreate) error {
	player := b.MusicPlayers.Get(*e.GuildID())
	strIndex := e.SlashCommandInteractionData().String("song")
	index, err := strconv.Atoi(strIndex)
	if err != nil {
		return e.CreateMessage(discord.MessageCreate{Content: p.Sprintf("modules.music.commands.remove.invalid.index")})
	}

	removeTrack := player.Queue.Get(index - 1)
	if removeTrack == nil {
		return e.CreateMessage(discord.MessageCreate{Content: p.Sprintf("modules.music.commands.remove.track.not.found", index)})
	}

	player.Queue.Remove(index - 1)
	return e.CreateMessage(discord.MessageCreate{
		Content: p.Sprintf("modules.music.commands.remove.removed", formatTrack(removeTrack), index),
	})
}

func removeUserSongsHandler(b *dbot.Bot, p *message.Printer, e *events.ApplicationCommandInteractionCreate) error {
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
		msg = p.Sprintf("modules.music.commands.remove.no.user.tracks", discord.UserMention(userID))
	} else {
		msg = p.Sprintf("modules.music.commands.remove.removed.user.tracks", removedTracks, discord.UserMention(userID))
	}

	return e.CreateMessage(discord.NewMessageCreateBuilder().
		SetContent(msg).
		SetAllowedMentions(&discord.AllowedMentions{}).
		Build(),
	)
}

func removeSongAutocompleteHandler(b *dbot.Bot, _ *message.Printer, e *events.AutocompleteInteractionCreate) error {
	player := b.MusicPlayers.Get(*e.GuildID())
	if player == nil || player.Queue.Len() == 0 {
		return e.Result(nil)
	}
	tracks := make([]string, player.Queue.Len())

	for i, track := range player.Queue.Tracks() {
		tracks[i] = fmt.Sprintf("%d. %s", i+1, track.Info().Title)
	}

	ranks := fuzzy.RankFindFold(e.Data.String("song"), tracks)

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
