package types

import (
	"net/http"

	"github.com/DisgoOrg/log"
	"github.com/uptrace/bun"
)

type Backend struct {
	Logger     log.Logger
	DB         *bun.DB
	HTTPServer *http.Server
	Config     Config
	Version    string
}
