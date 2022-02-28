package test

import (
	"github.com/DisgoOrg/disgo/core"
	"github.com/DisgoOrg/disgo/discord"
	"github.com/KittyBot-Org/KittyBotGo/internal/types"
)

var (
	_ types.Module         = (*Module)(nil)
	_ types.CommandsModule = (*Module)(nil)
	_ types.ListenerModule = (*Module)(nil)
)

type Module struct{}

func (Module) Commands() []types.Command {
	return []types.Command{
		{
			Create: discord.SlashCommandCreate{
				CommandName:       "test",
				Description:       "Test command",
				DefaultPermission: true,
			},
			CommandHandler: map[string]types.CommandHandler{
				"": testHandler,
			},
		},
	}
}

func (Module) OnEvent(b *types.Bot, event core.Event) {

}
