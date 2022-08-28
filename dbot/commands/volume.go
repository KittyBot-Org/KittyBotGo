package commands

import (
	"github.com/KittyBot-Org/KittyBotGo/dbot"
	"github.com/KittyBot-Org/KittyBotGo/dbot/responses"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/json"
	"golang.org/x/text/message"
)

var Volume = dbot.Command{
	Create: discord.SlashCommandCreate{
		Name:        "volume",
		Description: "Changes the volume of the music player.",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionInt{
				Name:        "volume",
				Description: "the desired volume",
				Required:    true,
				MinValue:    json.NewPtr(0),
				MaxValue:    json.NewPtr(100),
			},
		},
	},
	Checks: dbot.HasMusicPlayer.And(dbot.IsMemberConnectedToVoiceChannel),
	CommandHandler: map[string]dbot.CommandHandler{
		"": volumeHandler,
	},
}

func volumeHandler(b *dbot.Bot, p *message.Printer, e *events.ApplicationCommandInteractionCreate) error {
	player := b.MusicPlayers.Get(*e.GuildID())
	volume := e.SlashCommandInteractionData().Int("volume")
	if err := player.SetVolume(volume); err != nil {
		return e.CreateMessage(responses.CreateErrorf(p, "modules.music.commands.volume.error"))
	}
	return e.CreateMessage(responses.CreateSuccessf(p, "modules.music.commands.volume.success", volume))
}
