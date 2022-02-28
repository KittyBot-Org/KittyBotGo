package types

import (
	"context"

	"github.com/DisgoOrg/disgo/core"
	"github.com/DisgoOrg/disgo/core/bot"
	"github.com/DisgoOrg/disgo/discord"
	"github.com/DisgoOrg/disgo/gateway"
	"github.com/DisgoOrg/disgolink/disgolink"
	"github.com/DisgoOrg/log"
	"github.com/uptrace/bun"
)

type Bot struct {
	Logger           log.Logger
	Bot              *core.Bot
	Lavalink         disgolink.Link
	MusicPlayers     *MusicPlayerMap
	PlayHistoryCache *PlayHistoryCache
	Commands         *CommandMap
	Listeners        *Listeners
	DB               *bun.DB
	Config           Config
	Version          string
}

func (b *Bot) SetupBot() (err error) {
	b.Bot, err = bot.New(b.Config.Bot.Token,
		bot.WithLogger(b.Logger),
		bot.WithGatewayOpts(gateway.WithGatewayIntents(discord.GatewayIntentGuilds)),
		bot.WithEventListeners(b.Commands, b.Listeners),
	)
	return err
}

func (b *Bot) StartBot() (err error) {
	return b.Bot.ConnectGateway(context.TODO())
}
