package commands

import (
	"github.com/KittyBot-Org/KittyBotGo/dbot"
	"github.com/KittyBot-Org/KittyBotGo/dbot/responses"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgolink/lavalink"
	"golang.org/x/text/message"
)

var bassBoost = &lavalink.Equalizer{
	0:  0.2,
	1:  0.15,
	2:  0.1,
	3:  0.05,
	4:  0.0,
	5:  -0.05,
	6:  -0.1,
	7:  -0.1,
	8:  -0.1,
	9:  -0.1,
	10: -0.1,
	11: -0.1,
	12: -0.1,
	13: -0.1,
	14: -0.1,
}

var BassBoost = dbot.Command{
	Create: discord.SlashCommandCreate{
		Name:        "bass-boost",
		Description: "Enables or disables bass boost of the music player.",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionBool{
				Name:        "enable",
				Description: "if the bass boost should be enabled or disabled",
				Required:    true,
			},
		},
	},
	Checks: dbot.HasMusicPlayer.And(dbot.IsMemberConnectedToVoiceChannel),
	CommandHandler: map[string]dbot.CommandHandler{
		"": bassBoostHandler,
	},
}

func bassBoostHandler(b *dbot.Bot, p *message.Printer, e *events.ApplicationCommandInteractionCreate) error {
	player := b.MusicPlayers.Get(*e.GuildID())
	enable := e.SlashCommandInteractionData().Bool("enable")

	if enable {
		if err := player.Filters().SetEqualizer(bassBoost).Commit(); err != nil {
			return e.CreateMessage(responses.CreateErrorf(p, "modules.music.commands.bass.boost.enable.error"))
		}
		return e.CreateMessage(responses.CreateSuccessf(p, "modules.music.commands.bass.boost.enable.success"))
	}
	if err := player.Filters().SetEqualizer(&lavalink.Equalizer{}).Commit(); err != nil {
		return e.CreateMessage(responses.CreateErrorf(p, "modules.music.commands.bass.boost.disable.error"))
	}
	return e.CreateMessage(responses.CreateSuccessf(p, "modules.music.commands.bass.boost.disable.success"))
}
