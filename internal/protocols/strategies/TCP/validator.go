package TCP

import (
	"github.com/beelzebub-labs/beelzebub/v3/internal/parser"
)

type TCPValidator struct{}

func (v *TCPValidator) Name() string {
	return "tcp"
}

func (v *TCPValidator) Validate(config parser.BeelzebubServiceConfiguration) []parser.ValidationIssue {
	if config.Protocol != "tcp" {
		return nil
	}

	return parser.ValidateTLSConfig(config.TLSCertPath, config.TLSKeyPath)
}

func init() {
	parser.RegisterServiceValidator(&TCPValidator{})
}
