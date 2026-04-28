package SSH

import (
	"testing"

	"github.com/beelzebub-labs/beelzebub/v3/internal/parser"
	"github.com/stretchr/testify/assert"
)

func TestSSHValidator_Name(t *testing.T) {
	v := &SSHValidator{}
	assert.Equal(t, "ssh", v.Name())
}

func TestSSHValidator_NotSSHProtocol(t *testing.T) {
	v := &SSHValidator{}
	config := parser.BeelzebubServiceConfiguration{
		Protocol: "http",
	}
	issues := v.Validate(config)
	assert.Empty(t, issues)
}

func TestSSHValidator_EmptyPasswordRegex(t *testing.T) {
	v := &SSHValidator{}
	config := parser.BeelzebubServiceConfiguration{
		Protocol:               "ssh",
		PasswordRegex:          "",
		DeadlineTimeoutSeconds: 60,
	}
	issues := v.Validate(config)
	assert.Len(t, issues, 1)
	assert.Equal(t, parser.LevelError, issues[0].Level)
	assert.Equal(t, "passwordRegex is required for ssh protocol", issues[0].Message)
}

func TestSSHValidator_InvalidPasswordRegex(t *testing.T) {
	v := &SSHValidator{}
	config := parser.BeelzebubServiceConfiguration{
		Protocol:               "ssh",
		PasswordRegex:          "[",
		DeadlineTimeoutSeconds: 60,
	}
	issues := v.Validate(config)
	assert.Len(t, issues, 1)
	assert.Equal(t, parser.LevelError, issues[0].Level)
	assert.Contains(t, issues[0].Message, "passwordRegex is not a valid regex")
}

func TestSSHValidator_ValidPasswordRegex(t *testing.T) {
	v := &SSHValidator{}
	config := parser.BeelzebubServiceConfiguration{
		Protocol:               "ssh",
		PasswordRegex:          "^root$",
		DeadlineTimeoutSeconds: 60,
	}
	issues := v.Validate(config)
	assert.Empty(t, issues)
}

func TestSSHValidator_ZeroDeadline(t *testing.T) {
	v := &SSHValidator{}
	config := parser.BeelzebubServiceConfiguration{
		Protocol:               "ssh",
		PasswordRegex:          "^root$",
		DeadlineTimeoutSeconds: 0,
	}
	issues := v.Validate(config)
	assert.Empty(t, issues)
}

func TestSSHValidator_ValidDeadline(t *testing.T) {
	v := &SSHValidator{}
	config := parser.BeelzebubServiceConfiguration{
		Protocol:               "ssh",
		PasswordRegex:          "^root$",
		DeadlineTimeoutSeconds: 60,
	}
	issues := v.Validate(config)
	assert.Empty(t, issues)
}

func TestSSHValidator_MultipleIssues(t *testing.T) {
	v := &SSHValidator{}
	config := parser.BeelzebubServiceConfiguration{
		Protocol:      "ssh",
		PasswordRegex: "",
	}
	issues := v.Validate(config)
	assert.Len(t, issues, 1)
	assert.Equal(t, parser.LevelError, issues[0].Level)
	assert.Equal(t, "passwordRegex is required for ssh protocol", issues[0].Message)
}
