package historystore

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/beelzebub-labs/beelzebub/v3/internal/plugins"
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
	hs.Append("stale", plugins.Message{Role: "user", Content: "Hello"})
	hs.Append("fresh", plugins.Message{Role: "user", Content: "Hello"})

	// Back-date the stale entry beyond MaxHistoryAge.
	hs.Lock()
	e := hs.sessions["stale"]
	e.LastSeen = time.Now().Add(-MaxHistoryAge * 2)
	hs.sessions["stale"] = e
	hs.Unlock()

	CleanerInterval = 50 * time.Millisecond
	hs.HistoryCleaner()
	time.Sleep(200 * time.Millisecond)

	assert.False(t, hs.HasKey("stale"))
	assert.True(t, hs.HasKey("fresh"))
}

func TestAppend_Concurrent(t *testing.T) {
	hs := NewHistoryStore()
	const goroutines = 50
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func(i int) {
			defer wg.Done()
			hs.Append("key", plugins.Message{Role: "user", Content: fmt.Sprintf("msg%d", i)})
		}(i)
	}
	wg.Wait()
	assert.Len(t, hs.Query("key"), goroutines)
}

func TestQuery_MissingKey(t *testing.T) {
	hs := NewHistoryStore()
	assert.Nil(t, hs.Query("nonexistent"))
}
