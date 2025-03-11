package historystore

import (
	"sync"

	"github.com/mariocandela/beelzebub/v3/plugins"
)

// HistoryStore is a thread-safe structure for storing Messages used to build LLM Context.
type HistoryStore struct {
	sync.RWMutex
	sessions map[string][]plugins.Message
}

// NewHistoryStore returns a prepared HistoryStore
func NewHistoryStore() *HistoryStore {
	return &HistoryStore{
		sessions: make(map[string][]plugins.Message),
	}
}

// HasKey returns true if the supplied key exists in the map.
func (hs *HistoryStore) HasKey(key string) bool {
	hs.RLock()
	defer hs.RUnlock()
	_, ok := hs.sessions[key]
	return ok
}

// Query returns the value stored at the map
func (hs *HistoryStore) Query(key string) []plugins.Message {
	hs.RLock()
	defer hs.RUnlock()
	return hs.sessions[key]
}

// Append will add the slice of Mesages to the entry for the key.
// If the map has not yet been initalised, then a new map is created.
func (hs *HistoryStore) Append(key string, message ...plugins.Message) {
	hs.Lock()
	defer hs.Unlock()
	// In the unexpected case that the map has not yet been initalised, create it.
	if hs.sessions == nil {
		hs.sessions = make(map[string][]plugins.Message)
	}
	hs.sessions[key] = append(hs.sessions[key], message...)
}
