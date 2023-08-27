package tracer

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	mockStrategy := func(event Event) {}

	tracer := Init(mockStrategy)

	assert.NotNil(t, tracer.strategy)
}

func TestTraceEvent(t *testing.T) {
	eventCalled := Event{}
	var wg sync.WaitGroup

	mockStrategy := func(event Event) {
		defer wg.Done()

		eventCalled = event
	}

	tracer := Init(mockStrategy)

	wg.Add(1)
	tracer.TraceEvent(Event{
		ID:       "mockID",
		Protocol: HTTP.String(),
		Status:   Stateless.String(),
	})
	wg.Wait()

	assert.NotNil(t, eventCalled.ID)
	assert.Equal(t, "mockID", eventCalled.ID)
	assert.Equal(t, HTTP.String(), eventCalled.Protocol)
	assert.Equal(t, Stateless.String(), eventCalled.Status)
}

func TestSetStrategy(t *testing.T) {
	eventCalled := Event{}
	var wg sync.WaitGroup

	mockStrategy := func(event Event) {
		defer wg.Done()

		eventCalled = event
	}

	tracer := Init(mockStrategy)

	tracer.setStrategy(mockStrategy)

	wg.Add(1)
	tracer.TraceEvent(Event{
		ID:       "mockID",
		Protocol: HTTP.String(),
		Status:   Stateless.String(),
	})
	wg.Wait()

	assert.NotNil(t, eventCalled.ID)
	assert.Equal(t, "mockID", eventCalled.ID)
	assert.Equal(t, HTTP.String(), eventCalled.Protocol)
	assert.Equal(t, Stateless.String(), eventCalled.Status)
}

func TestStringStatus(t *testing.T) {
	assert.Equal(t, Start.String(), "Start")
	assert.Equal(t, End.String(), "End")
	assert.Equal(t, Stateless.String(), "Stateless")
	assert.Equal(t, Interaction.String(), "Interaction")
}
