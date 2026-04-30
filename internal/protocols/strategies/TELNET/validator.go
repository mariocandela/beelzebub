package TELNET

import (
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

	return parser.ValidatePasswordRegex(config.PasswordRegex, "telnet", config.Filename)
}

func init() {
	parser.RegisterServiceValidator(&TELNETValidator{})
}
