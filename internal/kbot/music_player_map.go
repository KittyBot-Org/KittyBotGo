package kbot

import (
	"sync"

	"github.com/disgoorg/snowflake"
)

func NewMusicPlayerMap(bot *Bot) *MusicPlayerMap {
	return &MusicPlayerMap{
		bot:     bot,
		players: make(map[snowflake.Snowflake]*MusicPlayer),
	}
}

type MusicPlayerMap struct {
	bot     *Bot
	mu      sync.Mutex
	players map[snowflake.Snowflake]*MusicPlayer
}

func (b *MusicPlayerMap) Len() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.players)
}

func (b *MusicPlayerMap) All() []*MusicPlayer {
	b.mu.Lock()
	defer b.mu.Unlock()

	players := make([]*MusicPlayer, len(b.players))
	i := 0
	for _, player := range b.players {
		players[i] = player
		i++
	}
	return players
}

func (b *MusicPlayerMap) Get(guildID snowflake.Snowflake) *MusicPlayer {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.players[guildID]
}

func (b *MusicPlayerMap) Has(guildID snowflake.Snowflake) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	_, ok := b.players[guildID]
	return ok
}

func (b *MusicPlayerMap) Add(player *MusicPlayer) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.players[player.GuildID()] = player
}

func (b *MusicPlayerMap) Delete(guildID snowflake.Snowflake) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.players, guildID)
}

func (b *MusicPlayerMap) New(guildID snowflake.Snowflake, playerType PlayerType, loopingType LoopingType) *MusicPlayer {
	player := &MusicPlayer{
		Player:    b.bot.Lavalink.Player(guildID),
		Bot:       b.bot,
		Type:      playerType,
		Queue:     NewMusicQueue(loopingType),
		History:   NewHistory(),
		SkipVotes: make(map[snowflake.Snowflake]struct{}),
	}
	player.AddListener(player)
	return player
}
