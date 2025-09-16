package tracer

import (
	"github.com/prometheus/client_golang/prometheus"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	mockStrategy := func(event Event) {}

	tracer := GetInstance(mockStrategy)

	assert.NotNil(t, tracer.strategy)
}

func TestTraceEvent(t *testing.T) {
	eventCalled := Event{}
	var wg sync.WaitGroup

	mockStrategy := func(event Event) {
		defer wg.Done()

		eventCalled = event
	}

	tracer := GetInstance(mockStrategy)

	tracer.strategy = mockStrategy

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

	tracer := GetInstance(mockStrategy)

	tracer.SetStrategy(mockStrategy)

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

type mockCounter struct {
	prometheus.Metric
	prometheus.Collector
	inc func()
	add func(float64)
}

var counter = 0

func (m mockCounter) Inc() {
	counter += 1
}

func (m mockCounter) Add(f float64) {
	counter = int(f)
}

func TestUpdatePrometheusCounters(t *testing.T) {
	mockStrategy := func(event Event) {}

	tracer := &tracer{
		strategy:        mockStrategy,
		eventsChan:      make(chan Event, Workers),
		eventsTotal:     mockCounter{},
		eventsSSHTotal:  mockCounter{},
		eventsTCPTotal:  mockCounter{},
		eventsHTTPTotal: mockCounter{},
		eventsMCPTotal:  mockCounter{},
	}

	tracer.updatePrometheusCounters(SSH.String())
	assert.Equal(t, 2, counter)

	tracer.updatePrometheusCounters(HTTP.String())
	assert.Equal(t, 4, counter)

	tracer.updatePrometheusCounters(TCP.String())
	assert.Equal(t, 6, counter)

	tracer.updatePrometheusCounters(MCP.String())
	assert.Equal(t, 8, counter)
}

func TestGetStrategy(t *testing.T) {
	mockStrategy := func(event Event) {}

	tracer := GetInstance(mockStrategy)

	retrievedStrategy := tracer.GetStrategy()
	assert.NotNil(t, retrievedStrategy)
}

func TestSetGetStrategyConcurrency(t *testing.T) {
	tracer := GetInstance(func(event Event) {})

	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(2)

		go func(id int) {
			defer wg.Done()
			mockStrategy := func(event Event) {}
			tracer.SetStrategy(mockStrategy)
		}(i)

		go func(id int) {
			defer wg.Done()
			strategy := tracer.GetStrategy()
			assert.NotNil(t, strategy)
		}(i)
	}

	wg.Wait()
}
