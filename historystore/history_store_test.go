package historystore

import (
	"testing"
	"time"

	"github.com/mariocandela/beelzebub/v3/plugins"
	"github.com/stretchr/testify/assert"
)

func TestNewHistoryStore(t *testing.T) {
	hs := NewHistoryStore()
	assert.NotNil(t, hs)
	assert.NotNil(t, hs.sessions)
}

func TestHasKey(t *testing.T) {
	hs := NewHistoryStore()
	hs.sessions["testKey"] = HistoryEvent{Messages: []plugins.Message{}}
	assert.True(t, hs.HasKey("testKey"))
	assert.False(t, hs.HasKey("nonExistentKey"))
}

func TestQuery(t *testing.T) {
	hs := NewHistoryStore()
	expectedMessages := []plugins.Message{{Role: "user", Content: "Hello"}}
	hs.sessions["testKey"] = HistoryEvent{Messages: expectedMessages}
	actualMessages := hs.Query("testKey")
	assert.Equal(t, expectedMessages, actualMessages)
}

func TestAppend(t *testing.T) {
	hs := NewHistoryStore()
	message1 := plugins.Message{Role: "user", Content: "Hello"}
	message2 := plugins.Message{Role: "assistant", Content: "Hi"}
	hs.Append("testKey", message1)
	assert.Equal(t, []plugins.Message{message1}, hs.sessions["testKey"].Messages)
	hs.Append("testKey", message2)
	assert.Equal(t, []plugins.Message{message1, message2}, hs.sessions["testKey"].Messages)
}

func TestAppendNilSessions(t *testing.T) {
	hs := &HistoryStore{}
	message1 := plugins.Message{Role: "user", Content: "Hello"}
	hs.Append("testKey", message1)
	assert.NotNil(t, hs.sessions)
	assert.Equal(t, []plugins.Message{message1}, hs.sessions["testKey"].Messages)
}

func TestHistoryCleaner(t *testing.T) {
	hs := NewHistoryStore()
	hs.Append("testKey", plugins.Message{Role: "user", Content: "Hello"})
	hs.Append("testKey2", plugins.Message{Role: "user", Content: "Hello"})

	// Make key older than MaxHistoryAge
	e := hs.sessions["testKey"]
	e.LastSeen = time.Now().Add(-MaxHistoryAge * 2)
	hs.sessions["testKey"] = e

	CleanerInterval = 5 * time.Second // Override for the test.
	hs.HistoryCleaner()
	time.Sleep(CleanerInterval + (1 * time.Second))

	assert.False(t, hs.HasKey("testKey"))
	assert.True(t, hs.HasKey("testKey2"))
}
