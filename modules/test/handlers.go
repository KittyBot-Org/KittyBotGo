package test

import (
	"github.com/DisgoOrg/disgo/core/events"
	"github.com/DisgoOrg/disgo/discord"
	"github.com/KittyBot-Org/KittyBotGo/internal/types"
	"golang.org/x/text/message"
)

func testHandler(b *types.Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) error {
	return e.CreateMessage(discord.NewMessageCreateBuilder().SetContent(p.Sprintf("commands.test", e.Member.EffectiveName(), e.Locale)).Build())
}
