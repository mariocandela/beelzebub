package SSH

import (
	"fmt"
	"regexp"

	"github.com/beelzebub-labs/beelzebub/v3/internal/parser"
)

type SSHValidator struct{}

func (v *SSHValidator) Name() string {
	return "ssh"
}

func (v *SSHValidator) Validate(config parser.BeelzebubServiceConfiguration) []parser.ValidationIssue {
	if config.Protocol != "ssh" {
		return nil
	}

	var issues []parser.ValidationIssue

	if config.PasswordRegex == "" {
		issues = append(issues, parser.ValidationIssue{
			Level:   parser.LevelError,
			Message: "passwordRegex is required for ssh protocol",
		})
	} else if _, err := regexp.Compile(config.PasswordRegex); err != nil {
		issues = append(issues, parser.ValidationIssue{
			Level:   parser.LevelError,
			Message: fmt.Sprintf("passwordRegex is not a valid regex: %v", err),
		})
	}

	return issues
}

func init() {
	parser.RegisterServiceValidator(&SSHValidator{})
}
