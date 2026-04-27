package TELNET

import (
	"fmt"
	"regexp"

	"github.com/beelzebub-labs/beelzebub/v3/internal/parser"
)

type TELNETValidator struct{}

func (v *TELNETValidator) Name() string {
	return "telnet"
}

func (v *TELNETValidator) Validate(config parser.BeelzebubServiceConfiguration) []parser.ValidationIssue {
	if config.Protocol != "telnet" {
		return nil
	}

	var issues []parser.ValidationIssue

	if config.PasswordRegex == "" {
		issues = append(issues, parser.ValidationIssue{
			Level:   parser.LevelError,
			Message: "passwordRegex is required for telnet protocol",
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
	parser.RegisterServiceValidator(&TELNETValidator{})
}
