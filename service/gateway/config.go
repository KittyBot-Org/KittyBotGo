package gateway

import (
	"fmt"
	"strings"

	"github.com/kittybot-org/kittybotgo/internal/log"
	"github.com/kittybot-org/kittybotgo/internal/otel"
)

type Config struct {
	Log     log.Config    `toml:"log"`
	Otel    otel.Config   `toml:"otel"`
	Gateway GatewayConfig `toml:"gateway"`
}

func (c Config) String() string {
	return fmt.Sprintf("\n Log: %v\n Otel: %s\n Gateway: %s\n",
		c.Log,
		c.Otel,
		c.Gateway,
	)
}

type GatewayConfig struct {
	GatewayURL string `toml:"gateway_url"`
	RestURL    string `toml:"rest_url"`
	Token      string `toml:"token"`
	NATSURL    string `toml:"nats_url"`
	Host       string `toml:"host"`
	Port       int    `toml:"port"`
}

func (c GatewayConfig) String() string {
	return fmt.Sprintf("\n  GatewayURL: %s\n  RestURL: %s\n  Token: %s\n  NATSURL: %s\n  Host: %s\n  Port: %d\n",
		c.GatewayURL,
		c.RestURL,
		strings.Repeat("*", len(c.Token)),
		c.NATSURL,
		c.Host,
		c.Port,
	)
}

func (c GatewayConfig) Address() string {
	host := c.Host
	if host == "" {
		host = "0.0.0.0"
	}
	port := c.Port
	if port == 0 {
		port = 8081
	}
	return fmt.Sprintf("%s:%d", host, port)
}
