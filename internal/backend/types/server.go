package types

import (
	"net/http"
)

func (b *Backend) SetupServer(handler http.Handler) {
	b.HTTPServer = &http.Server{
		Addr:    b.Config.Backend.Port,
		Handler: handler,
	}

	go func() {
		b.Logger.Info("Starting backend server on port: " + b.Config.Backend.Port)
		if err := b.HTTPServer.ListenAndServe(); err != nil {
			b.Logger.Fatal(err)
		}
	}()
}
