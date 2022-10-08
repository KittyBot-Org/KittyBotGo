package commands

import (
	"fmt"

	"github.com/KittyBot-Org/KittyBotGo/dbot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/handler"
)

func Loop(b *dbot.Bot) handler.Command {
	return handler.Command{
		Create: discord.SlashCommandCreate{
			Name:        "loop",
			Description: "Loops your queue.",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionInt{
					Name:        "looping-type",
					Description: "How to loop your queue",
					Required:    true,
					Choices: []discord.ApplicationCommandOptionChoiceInt{
						{
							Name:  "Off",
							Value: int(dbot.LoopingTypeOff),
						},
						{
							Name:  "Repeat Song",
							Value: int(dbot.LoopingTypeRepeatSong),
						},
						{
							Name:  "Repeat Queue",
							Value: int(dbot.LoopingTypeRepeatQueue),
						},
					},
				},
			},
		},
		Check: dbot.HasMusicPlayer(b).And(dbot.IsMemberConnectedToVoiceChannel(b)),
		CommandHandlers: map[string]handler.CommandHandler{
			"": loopHandler(b),
		},
	}
}

func loopHandler(b *dbot.Bot) handler.CommandHandler {
	return func(e *events.ApplicationCommandInteractionCreate) error {
		data := e.SlashCommandInteractionData()
		player := b.MusicPlayers.Get(*e.GuildID())
		loopingType := dbot.LoopingType(data.Int("looping-type"))
		player.Queue.SetType(loopingType)
		emoji := ""
		switch loopingType {
		case dbot.LoopingTypeRepeatSong:
			emoji = "🔂"
		case dbot.LoopingTypeRepeatQueue:
			emoji = "🔁"
		}
		return e.CreateMessage(discord.MessageCreate{
			Content: fmt.Sprintf("%s Looping: %s", emoji, loopingType),
		})
	}
}
