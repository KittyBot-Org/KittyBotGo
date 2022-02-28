package test

import (
	"github.com/DisgoOrg/disgo/core/events"
	"github.com/DisgoOrg/disgo/discord"
	"github.com/KittyBot-Org/KittyBotGo/internal/types"
)

func testHandler(b *types.Bot, e *events.ApplicationCommandInteractionEvent) error {
	return e.CreateMessage(discord.NewMessageCreateBuilder().SetContent("Test command").Build())
}
