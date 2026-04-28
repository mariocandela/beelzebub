package TELNET

import (
	"testing"

	"github.com/beelzebub-labs/beelzebub/v3/internal/parser"
	"github.com/stretchr/testify/assert"
)

func TestTELNETValidator_Name(t *testing.T) {
	v := &TELNETValidator{}
	assert.Equal(t, "telnet", v.Name())
}

func TestTELNETValidator_NotTELNETProtocol(t *testing.T) {
	v := &TELNETValidator{}
	config := parser.BeelzebubServiceConfiguration{
		Protocol: "ssh",
	}
	issues := v.Validate(config)
	assert.Empty(t, issues)
}

func TestTELNETValidator_EmptyPasswordRegex(t *testing.T) {
	v := &TELNETValidator{}
	config := parser.BeelzebubServiceConfiguration{
		Protocol: "telnet",
	}
	issues := v.Validate(config)
	assert.Len(t, issues, 1)
	assert.Equal(t, parser.LevelError, issues[0].Level)
	assert.Equal(t, "passwordRegex is required for telnet protocol", issues[0].Message)
}

func TestTELNETValidator_InvalidPasswordRegex(t *testing.T) {
	v := &TELNETValidator{}
	config := parser.BeelzebubServiceConfiguration{
		Protocol:      "telnet",
		PasswordRegex: "[",
	}
	issues := v.Validate(config)
	assert.Len(t, issues, 1)
	assert.Equal(t, parser.LevelError, issues[0].Level)
	assert.Contains(t, issues[0].Message, "passwordRegex is not a valid regex")
}

func TestTELNETValidator_ValidPasswordRegex(t *testing.T) {
	v := &TELNETValidator{}
	config := parser.BeelzebubServiceConfiguration{
		Protocol:      "telnet",
		PasswordRegex: "^root$",
	}
	issues := v.Validate(config)
	assert.Empty(t, issues)
}
