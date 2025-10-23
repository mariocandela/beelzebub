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

func TestAppendConversation(t *testing.T) {
	hs := NewHistoryStore()
	messageIn := plugins.Message{Role: "user", Content: "Hello"}
	messageOut := plugins.Message{Role: "assistant", Content: "Hi"}
	hs.AppendConversation("testKey", Conversation{Input: messageIn, Output: messageOut})
	assert.Equal(t, []plugins.Message{messageIn, messageOut}, hs.sessions["testKey"].Messages)
}

func TestAppendConversationDuplicates(t *testing.T) {
	hs := NewHistoryStore()
	messageIn := plugins.Message{Role: "user", Content: "Hello"}
	messageOut := plugins.Message{Role: "assistant", Content: "Hi"}
	conv := Conversation{Input: messageIn, Output: messageOut}
	hs.AppendConversation("testKey", conv)
	assert.Equal(t, []plugins.Message{messageIn, messageOut}, hs.sessions["testKey"].Messages)
	hs.AppendConversation("testKey", conv)
	// Both messages should be in the full Messages store twice, and only once in the Conversations store.
	assert.Equal(t, []plugins.Message{messageIn, messageOut, messageIn, messageOut}, hs.sessions["testKey"].Messages)
	assert.Equal(t, []Conversation{conv}, hs.sessions["testKey"].Conversations)
}

func TestAppendConversationNilSessions(t *testing.T) {
	hs := &HistoryStore{}
	messageIn := plugins.Message{Role: "user", Content: "Hello"}
	messageOut := plugins.Message{Role: "assistant", Content: "Hi"}
	conv := Conversation{Input: messageIn, Output: messageOut}
	hs.AppendConversation("testKey", conv)
	assert.NotNil(t, hs.sessions)
	assert.Equal(t, []plugins.Message{messageIn, messageOut}, hs.sessions["testKey"].Messages)
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

func TestQueryConversations(t *testing.T) {
	hs := NewHistoryStore()
	messageIn1 := plugins.Message{Role: "user", Content: "Hello"}
	messageOut1 := plugins.Message{Role: "assistant", Content: "Hi"}
	conv1 := Conversation{Input: messageIn1, Output: messageOut1}

	messageIn2 := plugins.Message{Role: "user", Content: "How are you?"}
	messageOut2 := plugins.Message{Role: "assistant", Content: "I am fine."}
	conv2 := Conversation{Input: messageIn2, Output: messageOut2}

	hs.sessions["testKey"] = HistoryEvent{
		Conversations: []Conversation{conv1, conv2},
	}

	t.Run("Conversation Found", func(t *testing.T) {
		foundConv := hs.QueryConversations("testKey", messageIn1)
		assert.NotNil(t, foundConv)
		assert.Equal(t, &conv1, foundConv)
	})

	t.Run("Conversation Not Found", func(t *testing.T) {
		notFoundConv := hs.QueryConversations("testKey", plugins.Message{Role: "user", Content: "non-existent"})
		assert.Nil(t, notFoundConv)
	})

	t.Run("Key Not Found", func(t *testing.T) {
		notFoundConv := hs.QueryConversations("nonExistentKey", messageIn1)
		assert.Nil(t, notFoundConv)
	})

	t.Run("Empty Conversations for Key", func(t *testing.T) {
		hs.sessions["emptyKey"] = HistoryEvent{Conversations: []Conversation{}}
		notFoundConv := hs.QueryConversations("emptyKey", messageIn1)
		assert.Nil(t, notFoundConv)
	})
}
