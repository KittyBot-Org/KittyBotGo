package commands

import (
	"github.com/KittyBot-Org/KittyBotGo/dbot"
	"github.com/KittyBot-Org/KittyBotGo/dbot/responses"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"golang.org/x/text/message"
)

var Pause = dbot.Command{
	Create: discord.SlashCommandCreate{
		CommandName: "pause",
		Description: "Pauses or resumes the music.",
	},
	Checks: dbot.HasMusicPlayer.And(dbot.IsMemberConnectedToVoiceChannel),
	CommandHandler: map[string]dbot.CommandHandler{
		"": pauseHandler,
	},
}

func pauseHandler(b *dbot.Bot, p *message.Printer, e *events.ApplicationCommandInteractionCreate) error {
	player := b.MusicPlayers.Get(*e.GuildID())
	pause := !player.Paused()
	if err := player.Pause(pause); err != nil {
		var msg string
		if pause {
			msg = p.Sprintf("modules.music.commands.pause.error")
		} else {
			msg = p.Sprintf("modules.music.commands.unpause.error")
		}
		return e.CreateMessage(responses.CreateSuccessf(p, msg))
	}
	var msg string
	if pause {
		msg = p.Sprintf("modules.music.commands.pause")
	} else {
		msg = p.Sprintf("modules.music.commands.unpause")
	}
	return e.CreateMessage(responses.CreateSuccessComponentsf(p, msg, nil, getMusicControllerComponents(player.PlayingTrack())))
}
