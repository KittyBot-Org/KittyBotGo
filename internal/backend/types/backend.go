package types

import (
	"net/http"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/KittyBot-Org/KittyBotGo/internal/bot/types"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/log"
	"github.com/procyon-projects/chrono"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/uptrace/bun"
)

type Backend struct {
	Logger        log.Logger
	DB            *bun.DB
	Rest          rest.Rest
	PrometheusAPI v1.API
	HTTPServer    *http.Server
	Scheduler     chrono.TaskScheduler
	Commands      []discord.ApplicationCommandCreate
	Config        Config
	Version       string

	PubSubClient *pubsub.Client
}

func (b *Backend) SetupRestServices() {
	rest.NewRest(rest.NewClient(b.Config.Token, rest.WithLogger(b.Logger)))
}

func (b *Backend) SetupPrometheusAPI() error {
	client, err := api.NewClient(api.Config{Address: b.Config.PrometheusEndpoint})
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
			commands := mod.Commands()
			cmds := make([]discord.ApplicationCommandCreate, len(commands))
			for i, cmd := range commands {
				cmds[i] = cmd.Create
			}
			b.Commands = append(b.Commands, cmds...)
		}
	}

	b.Logger.Infof("Loaded %d commands", len(b.Commands))
}
