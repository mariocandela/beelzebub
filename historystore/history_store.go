package historystore

import (
	"sync"
	"time"

	"github.com/mariocandela/beelzebub/v3/plugins"
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
	LastSeen      time.Time
	Messages      []plugins.Message
	Conversations []Conversation
}

// Conversation is a matching set of input and output messages.
type Conversation struct {
	Input, Output plugins.Message
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
// DEPRECATED: Use AppendConversation() instead.
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
	e.LastSeen = time.Now()
	e.Messages = append(e.Messages, message...)
	hs.sessions[key] = e
}

// AppendConverstion will update the entries to the stored messages and conversation cache.
// If the map has not yet been initalised, then a new map is created.
// Conversations are unique messages (each unique input is stored, otherwise discarded.)
// The messages are added to the full Message store even if they are not unique (avoiding the need to call Append as well).
func (hs *HistoryStore) AppendConverstion(key string, conversation Conversation) {
	hs.Lock()
	defer hs.Unlock()
	if hs.sessions == nil {
		hs.sessions = make(map[string]HistoryEvent)
	}
	e, ok := hs.sessions[key]
	if !ok {
		e = HistoryEvent{}
	}
	e.LastSeen = time.Now()
	e.Messages = append(e.Messages, conversation.Input, conversation.Output)

	for _, c := range e.Conversations {
		if c.Input.Content == conversation.Input.Content {
			// Already seen input, Update last seen and return.
			hs.sessions[key] = e
			return
		}
	}

	e.Conversations = append(e.Conversations, conversation)
	hs.sessions[key] = e
}

// QueryConversations searches the cached conversations to see if the provided input has already been seen.
// Returns nil if not found.
func (hs *HistoryStore) QueryConversations(key string, input plugins.Message) *Conversation {
	hs.RLock()
	defer hs.RUnlock()
	for _, c := range hs.sessions[key].Conversations {
		if c.Input.Content == input.Content {
			return &c
		}
	}
	return nil
}

// HistoryCleaner is a function that will periodically remove records from the HistoryStore
// that are older than MaxHistoryAge.
func (hs *HistoryStore) HistoryCleaner() {
	cleanerTicker := time.NewTicker(CleanerInterval)
	go func() {
		for range cleanerTicker.C {
			hs.Lock()
			for k, v := range hs.sessions {
				if time.Since(v.LastSeen) > MaxHistoryAge {
					delete(hs.sessions, k)
				}
			}
			hs.Unlock()
		}
	}()
}
