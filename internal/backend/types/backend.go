package types

import (
	"net/http"
	"time"

	"github.com/DisgoOrg/disgo/discord"
	"github.com/DisgoOrg/disgo/rest"
	"github.com/DisgoOrg/log"
	"github.com/KittyBot-Org/KittyBotGo/internal/bot/types"
	"github.com/procyon-projects/chrono"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/uptrace/bun"
)

type Backend struct {
	Logger        log.Logger
	DB            *bun.DB
	RestServices  rest.Services
	PrometheusAPI v1.API
	HTTPServer    *http.Server
	Scheduler     chrono.TaskScheduler
	Commands      []discord.ApplicationCommandCreate
	Config        Config
	Version       string
}

func (b *Backend) SetupRestServices() {
	config := &rest.DefaultConfig
	config.Logger = b.Logger
	config.BotTokenFunc = func() string { return b.Config.Bot.Token }
	rest.NewServices(rest.NewClient(config))
}

func (b *Backend) SetupPrometheusAPI() error {
	client, err := api.NewClient(api.Config{Address: b.Config.PrometheusAddress})
	if err != nil {
		return err
	}
	b.PrometheusAPI = v1.NewAPI(client)
	return nil
}

func (b *Backend) SetupScheduler() error {
	b.Scheduler = chrono.NewDefaultTaskScheduler()

	if _, err := b.Scheduler.ScheduleWithFixedDelay(b.VoteTask, time.Hour); err != nil {
		return err
	}
	return nil
}

func (b *Backend) LoadCommands(modules []types.Module) {
	b.Logger.Info("Loading commands...")

	for _, module := range modules {
		if mod, ok := module.(types.CommandsModule); ok {
			var cmds []discord.ApplicationCommandCreate
			for i, cmd := range mod.Commands() {
				cmds[i] = cmd.Create
			}
			b.Commands = append(b.Commands, cmds...)
		}
	}

	b.Logger.Infof("Loaded %d commands", len(b.Commands))
}
