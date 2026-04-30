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
	tests := []struct {
		user       string
		serverName string
		expected   string
	}{
		{"root", "ubuntu", "root@ubuntu:~$ "},
		{"admin", "debian", "admin@debian:~$ "},
		{"", "", "@:~$ "},
		{"user", "", "user@:~$ "},
		{"", "server", "@server:~$ "},
	}
	for _, tt := range tests {
		t.Run(tt.user+"@"+tt.serverName, func(t *testing.T) {
			assert.Equal(t, tt.expected, buildPrompt(tt.user, tt.serverName))
		})
	}
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

func TestSSHStrategy_Init_ReusesExistingSessions(t *testing.T) {
	strategy := &SSHStrategy{}
	mt := &mockTracer{}

	servConf := parser.BeelzebubServiceConfiguration{
		Address:                "127.0.0.1:0",
		DeadlineTimeoutSeconds: 1,
		PasswordRegex:          ".*",
	}

	assert.NoError(t, strategy.Init(servConf, mt))
	assert.NotNil(t, strategy.Sessions)

	original := strategy.Sessions

	// A second Init must reuse the same Sessions store, not replace it.
	assert.NoError(t, strategy.Init(servConf, mt))
	assert.Same(t, original, strategy.Sessions)
}

func TestSSHStrategy_Init_InvalidAddress(t *testing.T) {
	strategy := &SSHStrategy{}
	mt := &mockTracer{}

	servConf := parser.BeelzebubServiceConfiguration{
		Address:       "invalid-address-no-port",
		PasswordRegex: ".*",
	}

	// SSH runs the listener asynchronously; Init itself should not return an error.
	assert.NoError(t, strategy.Init(servConf, mt))
}
