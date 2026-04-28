package SSH

import (
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

	return parser.ValidatePasswordRegex(config.PasswordRegex, "ssh")
}

func init() {
	parser.RegisterServiceValidator(&SSHValidator{})
}
