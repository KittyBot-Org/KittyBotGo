package commands

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"

	"github.com/KittyBot-Org/KittyBotGo/service/bot"
)

func New(b *bot.Bot) *Cmds {
	cmds := &Cmds{
		Bot:    b,
		Router: handler.New(),
		Commands: []discord.ApplicationCommandCreate{
			ping,
			play,
			queue,
			playing,
			next,
			shuffle,
		},
	}
	cmds.HandleCommand("/ping", cmds.OnPing)
	cmds.HandleCommand("/play", cmds.OnPlay)
	cmds.Group(func(r handler.Router) {
		r.Use(cmds.OnHasPlayer)
		r.HandleCommand("/queue", cmds.OnQueue)
		r.HandleCommand("/playing", cmds.OnPlaying)
		r.HandleCommand("/next", cmds.OnNext)
		r.HandleCommand("/shuffle", cmds.OnShuffle)
	})
	return cmds
}

type Cmds struct {
	*bot.Bot
	handler.Router
	Commands []discord.ApplicationCommandCreate
}
