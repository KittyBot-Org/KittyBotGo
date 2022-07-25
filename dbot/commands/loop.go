package commands

import (
	"github.com/KittyBot-Org/KittyBotGo/dbot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"golang.org/x/text/message"
)

var Loop = handler.Command{
	Create: discord.SlashCommandCreate{
		CommandName: "loop",
		Description: "Loops your queue.",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionInt{
				OptionName:  "looping-type",
				Description: "how to loop your queue",
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
	Checks: dbot.HasMusicPlayer.And(dbot.IsMemberConnectedToVoiceChannel),
	CommandHandler: map[string]handler.CommandHandler{
		"": loopHandler,
	},
}

func loopHandler(b *dbot.Bot, p *message.Printer, e *events.ApplicationCommandInteractionCreate) error {
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
		Content: p.Sprintf("modules.commands.loop", emoji, loopingType),
	})
}
