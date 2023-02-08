package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/disgoorg/log"
)

func main() {
	cfgPath := flag.String("config", "config.json", "path to config.json")
	flag.Parse()

	logger := log.New(log.Ldate | log.Ltime | log.Lshortfile)
	logger.Infof("Backend is starting... (config path:%s)", *cfgPath)

	logger.Info("Backend is running. Press CTRL-C to exit.")
	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-s
}
