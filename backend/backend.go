package backend

import (
	"net/http"
	"time"

	"github.com/KittyBot-Org/KittyBotGo/db"
	"github.com/KittyBot-Org/KittyBotGo/dbot"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/log"
	"github.com/procyon-projects/chrono"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

type Backend struct {
	Logger        log.Logger
	DB            db.DB
	Rest          rest.Rest
	PrometheusAPI v1.API
	HTTPServer    *http.Server
	Scheduler     chrono.TaskScheduler
	Commands      []discord.ApplicationCommandCreate
	Config        Config
	Version       string
}

func (b *Backend) SetupRestServices() {
	rest.New(rest.NewClient(b.Config.Token, rest.WithLogger(b.Logger)))
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

func (b *Backend) LoadCommands(modules []dbot.Module) {
	b.Logger.Info("Loading commands...")

	for _, module := range modules {
		if mod, ok := module.(dbot.CommandsModule); ok {
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
