package handlers

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/lithammer/fuzzysearch/fuzzy"

	"github.com/KittyBot-Org/KittyBotGo/interal/database"
	"github.com/KittyBot-Org/KittyBotGo/service/bot/res"
)

var queueCommand = discord.SlashCommandCreate{
	Name:        "queue",
	Description: "Lets you manage the queue",
	Options: []discord.ApplicationCommandOption{
		discord.ApplicationCommandOptionSubCommand{
			Name:        "clear",
			Description: "Clears the queue",
		},
		discord.ApplicationCommandOptionSubCommand{
			Name:        "remove",
			Description: "Removes a track from the queue",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionInt{
					Name:         "track",
					Description:  "The track to remove",
					Required:     true,
					Autocomplete: true,
				},
			},
		},
		discord.ApplicationCommandOptionSubCommand{
			Name:        "shuffle",
			Description: "Shuffles the queue",
		},
		discord.ApplicationCommandOptionSubCommand{
			Name:        "show",
			Description: "Shows the queue",
		},
		discord.ApplicationCommandOptionSubCommand{
			Name:        "type",
			Description: "Lets you change the queue type",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionInt{
					Name:        "type",
					Description: "The type of queue",
					Required:    true,
					Choices: []discord.ApplicationCommandOptionChoiceInt{
						{
							Name:  "Normal",
							Value: int(database.QueueTypeNormal),
						},
						{
							Name:  "Repeat Track",
							Value: int(database.QueueTypeRepeatTrack),
						},
						{
							Name:  "Repeat Queue",
							Value: int(database.QueueTypeRepeatQueue),
						},
					},
				},
			},
		},
	},
}

func (h *Handlers) OnQueueShow(e *handler.CommandEvent) error {
	tracks, err := h.Database.GetQueue(*e.GuildID())
	if err != nil {
		return e.CreateMessage(res.CreateErr("Failed to get queue", err))
	}

	if len(tracks) == 0 {
		return e.CreateMessage(res.Create("The queue is empty"))
	}

	content := fmt.Sprintf("Queue(`%d`):\n", len(tracks))
	for i, track := range tracks {
		line := fmt.Sprintf("%d. %s\n", i+1, res.FormatTrack(track.Track, 0))
		if len([]rune(content))+len([]rune(line)) > 2000 {
			break
		}
		content += line
	}

	return e.CreateMessage(res.Create(content))
}

func (h *Handlers) OnQueueType(e *handler.CommandEvent) error {
	data := e.SlashCommandInteractionData()
	player := h.Lavalink.Player(*e.GuildID())
	queueType := database.QueueType(data.Int("type"))

	if err := h.Database.UpdatePlayer(database.Player{
		GuildID:   *e.GuildID(),
		Node:      player.Node().Config().Name,
		QueueType: queueType,
	}); err != nil {
		return e.CreateMessage(res.CreateErr("Failed to update player", err))
	}

	var emoji string
	switch queueType {
	case database.QueueTypeNormal:
		emoji = "➡️"
	case database.QueueTypeRepeatTrack:
		emoji = "🔂"
	case database.QueueTypeRepeatQueue:
		emoji = "🔁"
	}
	return e.CreateMessage(discord.MessageCreate{
		Content: fmt.Sprintf("%s Queuetype changed to: %s", emoji, queueType),
	})
}

func (h *Handlers) OnQueueShuffle(e *handler.CommandEvent) error {
	err := h.Database.ShuffleQueue(*e.GuildID())
	if errors.Is(err, sql.ErrNoRows) {
		return e.CreateMessage(res.CreateError("No more songs in queue"))
	}
	if err != nil {
		return e.CreateMessage(res.CreateErr("Failed to get next song", err))
	}

	return e.CreateMessage(res.Createf("Shuffled queue"))
}

func (h *Handlers) OnQueueClear(e *handler.CommandEvent) error {
	err := h.Database.ClearQueue(*e.GuildID())
	if err != nil {
		return e.CreateMessage(res.CreateErr("Failed to clear queue", err))
	}

	return e.CreateMessage(res.Createf("Cleared queue"))
}

func (h *Handlers) OnQueueRemove(e *handler.CommandEvent) error {
	data := e.SlashCommandInteractionData()
	trackID := data.Int("track")

	err := h.Database.RemoveQueueTrack(trackID)
	if errors.Is(err, sql.ErrNoRows) {
		return e.CreateMessage(res.CreateError("No more songs in queue"))
	}
	if err != nil {
		return e.CreateMessage(res.CreateErr("Failed to remove song", err))
	}

	return e.CreateMessage(res.Createf("Removed song from queue"))
}

func (h *Handlers) OnQueueAutocomplete(e *handler.AutocompleteEvent) error {
	tracks, err := h.Database.GetQueue(*e.GuildID())
	if err != nil {
		return e.Result(nil)
	}

	trackValues := make(map[string]int, len(tracks))
	trackNames := make([]string, len(tracks))
	for i, track := range tracks {
		name := trim(fmt.Sprintf("%d. %s - %s", i+1, track.Track.Info.Title, track.Track.Info.Author), 100)
		trackValues[name] = track.ID
		trackNames[i] = name
	}

	ranks := fuzzy.RankFindFold(e.Data.String("track"), trackNames)
	choicesLen := len(ranks)
	if choicesLen > 25 {
		choicesLen = 25
	}
	choices := make([]discord.AutocompleteChoice, choicesLen)
	for i, rank := range ranks {
		if i >= 25 {
			break
		}
		choices[i] = discord.AutocompleteChoiceInt{
			Name:  rank.Target,
			Value: trackValues[rank.Target],
		}
	}
	return e.Result(choices)
}

func trim(s string, length int) string {
	r := []rune(s)
	if len(r) > length {
		return string(r[:length-1]) + "…"
	}
	return s
}
