package kbot

import (
	"context"
	"sync"

	"github.com/disgoorg/disgolink/disgolink"
	"github.com/disgoorg/disgolink/lavalink"
	"github.com/disgoorg/source-plugins"
)

func (b *Bot) SetupLavalink() {
	b.MusicPlayers = NewMusicPlayerMap(b)
	b.Lavalink = disgolink.New(b.Client, lavalink.WithPlugins(source_plugins.NewSpotifyPlugin(), source_plugins.NewAppleMusicPlugin()))
	b.RegisterNodes()
}

func (b *Bot) RegisterNodes() {
	var wg sync.WaitGroup
	for i := range b.Config.Lavalink.Nodes {
		wg.Add(1)
		config := b.Config.Lavalink.Nodes[i]
		go func() {
			defer wg.Done()
			node, err := b.Lavalink.AddNode(context.TODO(), config)
			if err != nil {
				b.Logger.Error("Failed to add node: ", err)
				return
			}
			if config.ResumingKey != "" {
				if err = node.ConfigureResuming(config.ResumingKey, b.Config.Lavalink.ResumingTimeout); err != nil {
					b.Logger.Error("Failed to configure resuming: ", err)
				}
			}
		}()
	}
	wg.Wait()
}
