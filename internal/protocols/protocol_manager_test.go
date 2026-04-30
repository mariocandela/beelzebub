package protocols

import (
	"errors"
	"github.com/beelzebub-labs/beelzebub/v3/internal/parser"
	"github.com/beelzebub-labs/beelzebub/v3/internal/tracer"
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

func TestSetProtocolStrategy_ChangesStrategy(t *testing.T) {
	mockTraceStrategy := func(event tracer.Event) {}

	protocolManager := InitProtocolManager(mockTraceStrategy, mockServiceStrategyError{})

	// Error strategy initially returns an error.
	assert.Error(t, protocolManager.InitService(parser.BeelzebubServiceConfiguration{}))

	// After switching to the valid strategy, InitService should succeed.
	protocolManager.SetProtocolStrategy(mockServiceStrategyValid{})
	assert.NoError(t, protocolManager.InitService(parser.BeelzebubServiceConfiguration{}))
}

func TestInitProtocolManager_TracerNonNil(t *testing.T) {
	mockTraceStrategy := func(event tracer.Event) {}

	protocolManager := InitProtocolManager(mockTraceStrategy, mockServiceStrategyValid{})

	assert.NotNil(t, protocolManager.tracer)
	assert.NotNil(t, protocolManager.strategy)
}

func TestInitService_PassesConfigToStrategy(t *testing.T) {
	mockTraceStrategy := func(event tracer.Event) {}

	var captured parser.BeelzebubServiceConfiguration
	wantConf := parser.BeelzebubServiceConfiguration{Address: "0.0.0.0:8080", Protocol: "ssh"}

	rec := &recorderStrategy{captured: &captured}
	protocolManager := InitProtocolManager(mockTraceStrategy, rec)

	assert.NoError(t, protocolManager.InitService(wantConf))
	assert.Equal(t, wantConf, captured)
}

type recorderStrategy struct {
	captured *parser.BeelzebubServiceConfiguration
}

func (r *recorderStrategy) Init(conf parser.BeelzebubServiceConfiguration, _ tracer.Tracer) error {
	*r.captured = conf
	return nil
}
