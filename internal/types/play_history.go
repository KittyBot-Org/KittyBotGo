package types

import (
	"context"
	"sync"
	"time"

	"github.com/DisgoOrg/snowflake"
	"github.com/KittyBot-Org/KittyBotGo/internal/models"
)

const cacheLifetime = time.Minute * 5

type playHistoryEntry struct {
	history    []models.PlayHistory
	lastAccess time.Time
}

func NewPlayHistoryCache(bot *Bot) *PlayHistoryCache {
	cache := &PlayHistoryCache{
		bot:   bot,
		cache: make(map[snowflake.Snowflake]*playHistoryEntry),
	}
	cache.StartCleanup()
	return cache
}

type PlayHistoryCache struct {
	bot   *Bot
	mu    sync.Mutex
	cache map[snowflake.Snowflake]*playHistoryEntry
}

func (m *PlayHistoryCache) StartCleanup() {
	go func() {
		for range time.After(time.Minute * 1) {
			m.mu.Lock()
			for k, v := range m.cache {
				if time.Since(v.lastAccess) > cacheLifetime {
					delete(m.cache, k)
				}
			}
			m.mu.Unlock()
		}
	}()
}

func (m *PlayHistoryCache) Get(userID snowflake.Snowflake) ([]models.PlayHistory, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if v, ok := m.cache[userID]; ok {
		v.lastAccess = time.Now()
		return v.history, true
	}
	var history []models.PlayHistory
	if err := m.bot.DB.NewSelect().Model(&history).Where("user_id = ?", userID).Scan(context.TODO()); err != nil {
		m.bot.Logger.Error("Failed to get music history entries: ", err)
		return nil, false
	}
	m.cache[userID] = &playHistoryEntry{
		history:    history,
		lastAccess: time.Now(),
	}
	return history, true
}

func (m *PlayHistoryCache) Add(userID snowflake.Snowflake, query string, title string) {
	history := models.PlayHistory{
		UserID: userID,
		Query:  query,
		Title:  title,
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.cache[history.UserID]; ok {
		for _, entry := range m.cache[history.UserID].history {
			if entry.Query == history.Query {
				if _, err := m.bot.DB.NewUpdate().Model(&entry).WherePK().Exec(context.TODO()); err != nil {
					m.bot.Logger.Error("Failed to update music history entry: ", err)
				}
				return
			}
		}
		m.cache[history.UserID].history = append(m.cache[history.UserID].history, history)
	} else {
		m.cache[history.UserID] = &playHistoryEntry{
			history:    []models.PlayHistory{history},
			lastAccess: time.Now(),
		}
	}
	if _, err := m.bot.DB.NewInsert().Model(&history).Exec(context.TODO()); err != nil {
		m.bot.Logger.Error("Error adding music history entry: ", err)
	}
}
