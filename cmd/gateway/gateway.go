package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/disgoorg/log"

	"github.com/KittyBot-Org/KittyBotGo/config"
	"github.com/KittyBot-Org/KittyBotGo/service/gateway"
)

func main() {
	cfgPath := flag.String("config", "config.json", "path to config.json")
	flag.Parse()

	logger := log.New(log.Ldate | log.Ltime | log.Lshortfile)
	logger.Infof("Gateway is starting... (config path:%s)", *cfgPath)

	var cfg gateway.Config
	if err := config.Load(*cfgPath, &cfg); err != nil {
		logger.Fatalf("Failed to load config: %v", err)
	}

	gw, err := gateway.New(logger, cfg)
	if err != nil {
		logger.Fatalf("Failed to create gateway: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err = gw.Start(ctx); err != nil {
		logger.Fatalf("Failed to start gateway: %v", err)
	}

	logger.Info("Gateway is running. Press CTRL-C to exit.")
	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-s
}
