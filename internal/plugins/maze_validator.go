package plugins

import (
	"github.com/beelzebub-labs/beelzebub/v3/internal/parser"
)

type MazePluginValidator struct{}

func (v *MazePluginValidator) Name() string {
	return MazePluginName
}

func (v *MazePluginValidator) Validate(config parser.BeelzebubServiceConfiguration) []parser.ValidationIssue {
	if !usesPlugin(config, MazePluginName) {
		return nil
	}

	var issues []parser.ValidationIssue

	if config.Protocol != "http" {
		issues = append(issues, parser.ValidationIssue{
			Level:   parser.LevelWarning,
			Message: "plugin MazeHoneypot is only supported with http protocol",
		})
	}

	return issues
}

func init() {
	parser.RegisterServiceValidator(&MazePluginValidator{})
}
