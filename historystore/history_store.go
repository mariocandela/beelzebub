package historystore

import (
	"sync"
	"time"

	"github.com/mariocandela/beelzebub/v3/plugins"
	log "github.com/sirupsen/logrus"
)

var (
	MaxHistoryAge   = 60 * time.Minute
	CleanerInterval = 1 * time.Minute
)

// HistoryStore is a thread-safe structure for storing Messages used to build LLM Context.
type HistoryStore struct {
	sync.RWMutex
	sessions map[string]HistoryEvent
}

// HistoryEvent is a container for storing messages
type HistoryEvent struct {
	StartTime time.Time
	Messages  []plugins.Message
}

// NewHistoryStore returns a prepared HistoryStore
func NewHistoryStore() *HistoryStore {
	return &HistoryStore{
		sessions: make(map[string]HistoryEvent),
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
	return hs.sessions[key].Messages
}

// Append will add the slice of Mesages to the entry for the key.
// If the map has not yet been initalised, then a new map is created.
func (hs *HistoryStore) Append(key string, message ...plugins.Message) {
	hs.Lock()
	defer hs.Unlock()
	// In the unexpected case that the map has not yet been initalised, create it.
	if hs.sessions == nil {
		hs.sessions = make(map[string]HistoryEvent)
	}
	e, ok := hs.sessions[key]
	if !ok {
		e = HistoryEvent{}
	}
	e.StartTime = time.Now()
	e.Messages = append(e.Messages, message...)
	hs.sessions[key] = e
}

// HistoryCleaner is a function that will periodically remove records from the HistoryStore
// that are older than MaxHistoryAge.
func (hs *HistoryStore) HistoryCleaner() {
	cleanerTicker := time.NewTicker(CleanerInterval)
	go func() {
		for range cleanerTicker.C {
			hs.Lock()
			for k, v := range hs.sessions {
				if time.Since(v.StartTime) > MaxHistoryAge {
					log.Infof("removing key %q from history store", k)
					delete(hs.sessions, k)
				}
			}
			hs.Unlock()
		}
	}()
}
