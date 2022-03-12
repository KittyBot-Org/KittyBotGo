package types

import (
	"context"

	"github.com/DisgoOrg/disgo/core"
	"github.com/DisgoOrg/disgo/core/bot"
	"github.com/DisgoOrg/disgo/discord"
	"github.com/DisgoOrg/disgo/gateway"
	"github.com/DisgoOrg/disgolink/disgolink"
	"github.com/DisgoOrg/log"
	"github.com/DisgoOrg/utils/paginator"
	"github.com/uptrace/bun"
)

const KittyBotColor = 0x4c50c1

type Bot struct {
	Logger       log.Logger
	Bot          *core.Bot
	Lavalink     disgolink.Link
	MusicPlayers *MusicPlayerMap
	Paginator    *paginator.Manager
	Commands     *CommandMap
	Listeners    *Listeners
	DB           *bun.DB
	Config       Config
	Version      string
}

func (b *Bot) SetupPaginator() {
	b.Paginator = paginator.NewManager()
}

func (b *Bot) SetupBot() (err error) {
	b.Bot, err = bot.New(b.Config.Token,
		bot.WithLogger(b.Logger),
		bot.WithGatewayOpts(gateway.WithGatewayIntents(discord.GatewayIntentGuilds, discord.GatewayIntentGuildVoiceStates)),
		bot.WithEventListeners(b.Commands, b.Paginator, b.Listeners),
	)
	return err
}

func (b *Bot) StartBot() (err error) {
	return b.Bot.ConnectGateway(context.TODO())
}
