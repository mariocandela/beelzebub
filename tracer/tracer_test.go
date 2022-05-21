package tracer

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInit(t *testing.T) {
	mockStrategy := func(event Event) {}

	tracer := Init(mockStrategy)

	assert.NotNil(t, tracer.strategy)
}

func TestTraceEvent(t *testing.T) {
	eventCalled := Event{}

	mockStrategy := func(event Event) {
		eventCalled = event
	}

	tracer := Init(mockStrategy)

	tracer.TraceEvent(Event{
		ID:       "mockID",
		Protocol: HTTP.String(),
		Status:   Stateless.String(),
	})

	assert.NotNil(t, eventCalled.ID)
	assert.Equal(t, eventCalled.ID, "mockID")
	assert.Equal(t, eventCalled.Protocol, HTTP.String())
	assert.Equal(t, eventCalled.Status, Stateless.String())
}

func TestStringStatus(t *testing.T) {
	assert.Equal(t, Start.String(), "Start")
	assert.Equal(t, End.String(), "End")
	assert.Equal(t, Stateless.String(), "Stateless")
	assert.Equal(t, Interaction.String(), "Interaction")
}
