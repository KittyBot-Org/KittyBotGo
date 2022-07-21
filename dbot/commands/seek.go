package commands

import (
	"github.com/KittyBot-Org/KittyBotGo/dbot"
	"github.com/KittyBot-Org/KittyBotGo/dbot/responses"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/json"
	"github.com/disgoorg/disgolink/lavalink"
	"golang.org/x/text/message"
)

var Seek = dbot.Command{
	Create: discord.SlashCommandCreate{
		CommandName: "seek",
		Description: "Seeks the music to a point in the queue.",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionInt{
				OptionName:  "position",
				Description: "the position to seek to in seconds(default)/minutes/hours",
				Required:    true,
				MinValue:    json.NewPtr(0),
			},
			discord.ApplicationCommandOptionInt{
				OptionName:  "time-unit",
				Description: "in which time unit to seek",
				Required:    false,
				Choices: []discord.ApplicationCommandOptionChoiceInt{
					{
						Name:  "Seconds",
						Value: int(lavalink.Second),
					},
					{
						Name:  "Minutes",
						Value: int(lavalink.Minute),
					},
					{
						Name:  "Hours",
						Value: int(lavalink.Hour),
					},
				},
			},
		},
	},
	Checks: dbot.HasMusicPlayer.And(dbot.IsMemberConnectedToVoiceChannel),
	CommandHandler: map[string]dbot.CommandHandler{
		"": seekHandler,
	},
}

func seekHandler(b *dbot.Bot, p *message.Printer, e *events.ApplicationCommandInteractionCreate) error {
	data := e.SlashCommandInteractionData()
	player := b.MusicPlayers.Get(*e.GuildID())
	position := data.Int("position")
	timeUnit := lavalink.Second
	if timeUnitPtr, ok := data.OptInt("time-unit"); ok {
		timeUnit = lavalink.Duration(timeUnitPtr)
	}

	finalPosition := lavalink.Duration(position) * timeUnit
	if finalPosition > player.PlayingTrack().Info().Length {
		return e.CreateMessage(responses.CreateErrorf(p, "modules.music.commands.seek.position.too.big"))
	}
	if err := player.Seek(finalPosition); err != nil {
		return e.CreateMessage(responses.CreateErrorf(p, "modules.music.commands.seek.error"))
	}
	return e.CreateMessage(responses.CreateSuccessf(p, "modules.music.commands.seek.success"))
}
