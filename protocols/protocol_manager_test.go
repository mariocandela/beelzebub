package protocols

import (
	"errors"
	"github.com/mariocandela/beelzebub/v3/parser"
	"github.com/mariocandela/beelzebub/v3/tracer"
	"github.com/stretchr/testify/assert"
	"testing"
)

type mockServiceStrategyValid struct {
}

func (mockServiceStrategy mockServiceStrategyValid) Init(beelzebubServiceConfiguration parser.BeelzebubServiceConfiguration, tr tracer.Tracer) error {
	return nil
}

type mockServiceStrategyError struct {
}

func (mockServiceStrategy mockServiceStrategyError) Init(beelzebubServiceConfiguration parser.BeelzebubServiceConfiguration, tr tracer.Tracer) error {
	return errors.New("mockError")
}

func TestInitServiceManager(t *testing.T) {
	mockTraceStrategy := func(event tracer.Event) {}

	protocolManager := InitProtocolManager(mockTraceStrategy, mockServiceStrategyValid{})

	assert.NotNil(t, protocolManager.strategy)
	assert.NotNil(t, protocolManager.tracer)
}

func TestInitServiceSuccess(t *testing.T) {
	mockTraceStrategy := func(event tracer.Event) {}

	protocolManager := InitProtocolManager(mockTraceStrategy, mockServiceStrategyValid{})

	protocolManager.SetProtocolStrategy(mockServiceStrategyValid{})

	assert.Nil(t, protocolManager.InitService(parser.BeelzebubServiceConfiguration{}))
}

func TestInitServiceError(t *testing.T) {
	mockTraceStrategy := func(event tracer.Event) {}

	protocolManager := InitProtocolManager(mockTraceStrategy, mockServiceStrategyError{})

	assert.NotNil(t, protocolManager.InitService(parser.BeelzebubServiceConfiguration{}))
}
