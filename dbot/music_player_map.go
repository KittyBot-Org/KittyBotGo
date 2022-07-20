package dbot

import (
	"sync"

	"github.com/disgoorg/snowflake/v2"
)

func NewMusicPlayerMap(bot *Bot) *MusicPlayerMap {
	return &MusicPlayerMap{
		bot:     bot,
		players: make(map[snowflake.ID]*MusicPlayer),
	}
}

type MusicPlayerMap struct {
	bot     *Bot
	mu      sync.Mutex
	players map[snowflake.ID]*MusicPlayer
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

func (b *MusicPlayerMap) Get(guildID snowflake.ID) *MusicPlayer {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.players[guildID]
}

func (b *MusicPlayerMap) Has(guildID snowflake.ID) bool {
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

func (b *MusicPlayerMap) Delete(guildID snowflake.ID) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.players, guildID)
}

func (b *MusicPlayerMap) New(guildID snowflake.ID, playerType PlayerType, loopingType LoopingType) *MusicPlayer {
	player := &MusicPlayer{
		Player:    b.bot.Lavalink.Player(guildID),
		Bot:       b.bot,
		Type:      playerType,
		Queue:     NewMusicQueue(loopingType),
		History:   NewHistory(),
		SkipVotes: make(map[snowflake.ID]struct{}),
	}
	player.AddListener(player)
	return player
}
