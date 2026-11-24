package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/topi314/tint"

	"github.com/kittybot-org/kittybotgo/internal/config"
	"github.com/kittybot-org/kittybotgo/internal/log"
	"github.com/kittybot-org/kittybotgo/internal/otel"
	"github.com/kittybot-org/kittybotgo/service/gateway"
)

func main() {
	cfgPath := flag.String("config", "config.toml", "path to config file")
	flag.Parse()

	slog.Info("Gateway is starting...", slog.String("config", *cfgPath))

	var cfg gateway.Config
	if err := config.Load(*cfgPath, &cfg); err != nil {
		slog.Error("failed to load config", tint.Err(err))
		return
	}
	slog.Info("Config loaded", slog.String("config", cfg.String()))
	log.Setup(cfg.Log)

	shutdown, err := otel.Setup("kittybot-gateway", "dev", cfg.Otel)
	if err != nil {
		slog.Error("failed to setup otel", tint.Err(err))
		return
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := shutdown(ctx); err != nil {
			slog.Error("failed to shutdown otel", tint.Err(err))
		}
	}()

	g := gateway.New(cfg)
	defer g.Close()

	if err := g.Start(); err != nil {
		slog.Error("failed to start gateway", tint.Err(err))
		return
	}

	slog.Info("Gateway is running. Press CTRL-C to exit.", slog.String("cache", fmt.Sprintf("http://%s", cfg.Gateway.Address())))
	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM)
	<-s
}
