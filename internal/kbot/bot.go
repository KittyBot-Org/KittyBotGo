package kbot

import (
	"context"

	"github.com/KittyBot-Org/KittyBotGo/internal/db"
	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/disgolink/disgolink"
	"github.com/disgoorg/log"
	"github.com/disgoorg/utils/paginator"
)

type Bot struct {
	Logger       log.Logger
	Client       bot.Client
	Lavalink     disgolink.Link
	MusicPlayers *MusicPlayerMap
	Paginator    *paginator.Manager
	Commands     *CommandMap
	Listeners    *Listeners
	DB           db.DB
	Config       Config
	Version      string
}

func (b *Bot) SetupPaginator() {
	b.Paginator = paginator.NewManager()
}

func (b *Bot) SetupBot() (err error) {
	b.Client, err = disgo.New(b.Config.Token,
		bot.WithLogger(b.Logger),
		bot.WithGatewayConfigOpts(gateway.WithIntents(gateway.IntentGuilds, gateway.IntentGuildVoiceStates)),
		bot.WithEventListeners(b.Commands, b.Paginator, b.Listeners),
	)
	return err
}

func (b *Bot) StartBot() (err error) {
	return b.Client.ConnectGateway(context.TODO())
}
