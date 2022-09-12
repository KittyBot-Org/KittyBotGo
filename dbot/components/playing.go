package components

import (
	"github.com/KittyBot-Org/KittyBotGo/dbot"
	"github.com/KittyBot-Org/KittyBotGo/dbot/responses"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/handler"
	"golang.org/x/text/message"
)

func Previous(b *dbot.Bot) handler.Component {
	return handler.Component{
		Name:    "previous",
		Check:   nil,
		Handler: nil,
	}
}

func previousComponentHandler(b *dbot.Bot, _ []string, p *message.Printer, e *events.ComponentInteractionCreate) error {
	player, err := checkPlayer(b, p, e)
	if player == nil {
		return err
	}
	nextTrack := player.History.Last()
	if nextTrack == nil {
		return e.CreateMessage(responses.CreateErrorf(p, "modules.music.components.previous.empty"))
	}

	if err = player.Play(nextTrack); err != nil {
		return e.CreateMessage(responses.CreateErrorf(p, "modules.music.components.previous.error"))
	}
	return e.UpdateMessage(responses.UpdateSuccessComponentsf(p, "modules.music.commands.previous.success", []any{formatTrack(nextTrack), nextTrack.Info().Length}, getMusicControllerComponents(nextTrack)))
}
