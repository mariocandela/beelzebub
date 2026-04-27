package SSH

import (
	"testing"

	"github.com/beelzebub-labs/beelzebub/v3/internal/parser"
	"github.com/beelzebub-labs/beelzebub/v3/internal/tracer"
	"github.com/stretchr/testify/assert"
)

type mockTracer struct {
	events []tracer.Event
}

func (m *mockTracer) TraceEvent(event tracer.Event) {
	m.events = append(m.events, event)
}

func TestBuildPrompt(t *testing.T) {
	result := buildPrompt("root", "ubuntu")
	assert.Equal(t, "root@ubuntu:~$ ", result)
}

func TestBuildPrompt_EmptyFields(t *testing.T) {
	result := buildPrompt("", "")
	assert.Equal(t, "@:~$ ", result)
}

func TestSSHStrategy_Init_InvalidAddress(t *testing.T) {
	strategy := &SSHStrategy{}
	mt := &mockTracer{}

	servConf := parser.BeelzebubServiceConfiguration{
		Address:       "invalid-address-no-port",
		PasswordRegex: ".*",
	}

	err := strategy.Init(servConf, mt)
	// SSH should not return an error for invalid address at init time (it runs asynchronously)
	// but at least exercise the Init code path
	assert.NoError(t, err)
}

func TestSSHStrategy_Init_ValidAddress(t *testing.T) {
	strategy := &SSHStrategy{}
	mt := &mockTracer{}

	servConf := parser.BeelzebubServiceConfiguration{
		Address:                "127.0.0.1:0",
		Description:            "test SSH",
		DeadlineTimeoutSeconds: 2,
		PasswordRegex:          ".*",
	}

	err := strategy.Init(servConf, mt)
	assert.NoError(t, err)
	assert.NotNil(t, strategy.Sessions)
}
