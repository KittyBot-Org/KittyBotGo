package dbot

import (
	"context"

	"github.com/KittyBot-Org/KittyBotGo/db"
	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/disgo/webhook"
	"github.com/disgoorg/disgolink/disgolink"
	"github.com/disgoorg/log"
	"github.com/disgoorg/snowflake/v2"
	"github.com/disgoorg/utils/paginator"
)

func New(logger log.Logger, config Config, version string) *Bot {
	return &Bot{
		Logger:            logger,
		Paginator:         paginator.NewManager(),
		ReportLogWebhooks: map[snowflake.ID]webhook.Client{},
		Config:            config,
		Version:           version,
	}
}

type Bot struct {
	Logger            log.Logger
	Client            bot.Client
	Lavalink          disgolink.Link
	MusicPlayers      *MusicPlayerMap
	Paginator         *paginator.Manager
	CommandMap        *CommandMap
	DB                db.DB
	ReportLogWebhooks map[snowflake.ID]webhook.Client
	Config            Config
	Version           string
}

func (b *Bot) SetupBot(listeners ...bot.EventListener) (err error) {
	b.Client, err = disgo.New(b.Config.Token,
		bot.WithLogger(b.Logger),
		bot.WithGatewayConfigOpts(gateway.WithIntents(gateway.IntentGuilds, gateway.IntentGuildVoiceStates)),
		bot.WithEventListeners(append([]bot.EventListener{b.CommandMap, b.Paginator}, listeners...)...),
	)
	return err
}

func (b *Bot) StartBot() error {
	return b.Client.OpenGateway(context.TODO())
}
