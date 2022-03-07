package types

import (
	"net/http"

	"github.com/KittyBot-Org/KittyBotGo/internal/backend/routes"
)

func (b *Backend) SetupServer() {
	b.HTTPServer = &http.Server{
		Addr:    b.Config.Backend.Port,
		Handler: routes.Handler(b),
	}

	go func() {
		b.Logger.Info("Starting backend server on port: " + b.Config.Backend.Port)
		if err := b.HTTPServer.ListenAndServe(); err != nil {
			b.Logger.Fatal(err)
		}
	}()
}
