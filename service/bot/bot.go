package bot

import (
	"fmt"

	"github.com/disgoorg/log"
	"github.com/nats-io/nats.go"
)

func New(logger log.Logger, cfg Config) (*Bot, error) {
	b := &Bot{
		Config: cfg,
		Logger: logger,
	}

	conn, err := nats.Connect(cfg.Nats.URL,
		nats.Name("gateway"),
		nats.UserInfo(cfg.Nats.User, cfg.Nats.Password),
		nats.MaxReconnects(-1),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to nats: %w", err)
	}

	b.Nats = conn
	return b, nil
}

type Bot struct {
	Config Config
	Logger log.Logger
	Nats   *nats.Conn
}

func (b *Bot) Start() error {
	_, err := b.Nats.Subscribe("gateway.events.*", func(msg *nats.Msg) {
		msg.Ack()
		fmt.Println(msg.Subject, string(msg.Data))
	})
	return err
}

func (b *Bot) Close() {
	_ = b.Nats.Drain()
	b.Nats.Close()
}
