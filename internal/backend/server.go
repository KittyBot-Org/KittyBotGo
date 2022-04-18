package backend

import (
	"net/http"
)

func (b *Backend) SetupServer(handler http.Handler) {
	b.HTTPServer = &http.Server{
		Addr:    b.Config.Address,
		Handler: handler,
	}

	go func() {
		b.Logger.Info("Starting backend server on port: " + b.Config.Address)
		if err := b.HTTPServer.ListenAndServe(); err != nil {
			b.Logger.Fatal(err)
		}
	}()
}
