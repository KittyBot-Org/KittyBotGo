package types

import (
	"net/http"

	"github.com/DisgoOrg/disgo/rest"
	"github.com/DisgoOrg/log"
	"github.com/uptrace/bun"
)

type Backend struct {
	Logger       log.Logger
	DB           *bun.DB
	RestServices rest.Services
	HTTPServer   *http.Server
	Config       Config
	Version      string
}

func (b *Backend) SetupRestServices() {
	config := &rest.DefaultConfig
	config.Logger = b.Logger
	config.BotTokenFunc = func() string { return b.Config.Bot.Token }
	rest.NewServices(rest.NewClient(config))
}
