package plugins

import (
	"testing"

	"github.com/beelzebub-labs/beelzebub/v3/internal/parser"
	"github.com/stretchr/testify/assert"
)

func makeMazeService(protocol string) parser.BeelzebubServiceConfiguration {
	return parser.BeelzebubServiceConfiguration{
		Protocol: protocol,
		Commands: []parser.Command{
			{RegexStr: "^(.+)$", Plugin: MazePluginName},
		},
	}
}

func TestMazePluginValidator_NotUsed(t *testing.T) {
	validator := &MazePluginValidator{}

	config := parser.BeelzebubServiceConfiguration{
		Commands: []parser.Command{
			{RegexStr: "^(.+)$", Plugin: "SomeOtherPlugin"},
		},
	}

	issues := validator.Validate(config)
	assert.Empty(t, issues)
}

func TestMazePluginValidator_HTTPProtocol(t *testing.T) {
	validator := &MazePluginValidator{}

	config := makeMazeService("http")

	issues := validator.Validate(config)
	assert.Empty(t, issues)
}

func TestMazePluginValidator_SSHProtocol(t *testing.T) {
	validator := &MazePluginValidator{}

	config := makeMazeService("ssh")

	issues := validator.Validate(config)
	assert.Len(t, issues, 1)
	assert.Equal(t, parser.LevelWarning, issues[0].Level)
	assert.Contains(t, issues[0].Message, "only supported with http protocol")
}

func TestMazePluginValidator_TelnetProtocol(t *testing.T) {
	validator := &MazePluginValidator{}

	config := makeMazeService("telnet")

	issues := validator.Validate(config)
	assert.Len(t, issues, 1)
	assert.Equal(t, parser.LevelWarning, issues[0].Level)
	assert.Contains(t, issues[0].Message, "only supported with http protocol")
}

func TestMazePluginValidator_EmptyProtocol(t *testing.T) {
	validator := &MazePluginValidator{}

	config := makeMazeService("")

	issues := validator.Validate(config)
	assert.Len(t, issues, 1)
	assert.Equal(t, parser.LevelWarning, issues[0].Level)
	assert.Contains(t, issues[0].Message, "only supported with http protocol")
}
