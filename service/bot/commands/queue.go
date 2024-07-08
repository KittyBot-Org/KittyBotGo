package commands

import (
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/disgolink/v3/lavalink"
	"github.com/disgoorg/lavaqueue-plugin"
	"go.gopad.dev/fuzzysearch/fuzzy"

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
			Description: "Removes a song from the queue",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionInt{
					Name:         "song",
					Description:  "The song to remove",
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
				discord.ApplicationCommandOptionString{
					Name:        "type",
					Description: "The type of queue",
					Required:    true,
					Choices: []discord.ApplicationCommandOptionChoiceString{
						{
							Name:  "Normal",
							Value: string(lavaqueue.QueueTypeNormal),
						},
						{
							Name:  "Repeat Track",
							Value: string(lavaqueue.QueueTypeRepeatTrack),
						},
						{
							Name:  "Repeat Queue",
							Value: string(lavaqueue.QueueTypeRepeatQueue),
						},
					},
				},
			},
		},
	},
}

func (c *commands) OnQueueShow(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	player := c.Lavalink.ExistingPlayer(*e.GuildID())
	queue, err := lavaqueue.GetQueue(e.Ctx, player.Node(), *e.GuildID())
	if err != nil {
		return e.CreateMessage(res.CreateErr("Failed to get queue", err))
	}

	if len(queue.Tracks) == 0 {
		return e.CreateMessage(res.Create("The queue is empty"))
	}

	content := fmt.Sprintf("Queue(`%d`):\n", len(queue.Tracks))
	for i, track := range queue.Tracks {
		line := fmt.Sprintf("%d. %s\n", i+1, res.FormatTrack(track, 0))
		if len([]rune(content))+len([]rune(line)) > 2000 {
			break
		}
		content += line
	}

	return e.CreateMessage(res.Create(content))
}

func (c *commands) OnQueueType(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	player := c.Lavalink.Player(*e.GuildID())
	queueType := lavaqueue.QueueType(data.String("type"))

	if _, err := lavaqueue.UpdateQueue(e.Ctx, player.Node(), *e.GuildID(), lavaqueue.QueueUpdate{
		Type: &queueType,
	}); err != nil {
		return e.CreateMessage(res.CreateErr("Failed to update player", err))
	}

	var emoji string
	switch queueType {
	case lavaqueue.QueueTypeNormal:
		emoji = "➡️"
	case lavaqueue.QueueTypeRepeatTrack:
		emoji = "🔂"
	case lavaqueue.QueueTypeRepeatQueue:
		emoji = "🔁"
	}
	return e.CreateMessage(discord.MessageCreate{
		Content: fmt.Sprintf("%s Queuetype changed to: %s", emoji, queueType),
	})
}

func (c *commands) OnQueueShuffle(_ discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	player := c.Lavalink.Player(*e.GuildID())
	if err := lavaqueue.ShuffleQueue(e.Ctx, player.Node(), *e.GuildID()); err != nil {
		return e.CreateMessage(res.CreateErr("Failed to shuffle queue", err))
	}

	return e.CreateMessage(res.Createf("🔀 Shuffled queue"))
}

func (c *commands) OnQueueClear(_ discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	player := c.Lavalink.Player(*e.GuildID())
	if err := lavaqueue.ClearQueue(e.Ctx, player.Node(), *e.GuildID()); err != nil {
		return e.CreateMessage(res.CreateErr("Failed to clear queue", err))
	}

	return e.CreateMessage(res.Createf("🧹 Cleared queue"))
}

func (c *commands) OnQueueRemove(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	trackID := data.Int("song")

	player := c.Lavalink.Player(*e.GuildID())
	if err := lavaqueue.RemoveQueueTrack(e.Ctx, player.Node(), *e.GuildID(), trackID); err != nil {
		return e.CreateMessage(res.CreateErr("Failed to remove song", err))
	}

	return e.CreateMessage(res.Createf("Removed song from queue"))
}

type Track lavalink.Track

func (t Track) FilterValue() string {
	return t.Encoded
}

func (c *commands) OnQueueAutocomplete(e *handler.AutocompleteEvent) error {
	player := c.Lavalink.ExistingPlayer(*e.GuildID())
	queue, err := lavaqueue.GetQueue(e.Ctx, player.Node(), *e.GuildID())
	if err != nil {
		return e.AutocompleteResult(nil)
	}

	tracks := make([]Track, len(queue.Tracks))
	for i, track := range queue.Tracks {
		tracks[i] = Track(track)
	}

	ranks := fuzzy.RankFindFold(e.Data.String("query"), tracks)

	choices := make([]discord.AutocompleteChoice, len(ranks))
	for i, rank := range ranks {
		choices[i] = discord.AutocompleteChoiceInt{
			Name:  res.Trim(fmt.Sprintf("%d. %s", i+1, rank.Target.Info.Title), 100),
			Value: rank.OriginalIndex,
		}
	}
	return e.AutocompleteResult(choices)
}
