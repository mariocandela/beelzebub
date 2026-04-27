package TCP

import (
	"os"

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

	var issues []parser.ValidationIssue

	if (config.TLSCertPath != "" && config.TLSKeyPath == "") || (config.TLSCertPath == "" && config.TLSKeyPath != "") {
		issues = append(issues, parser.ValidationIssue{
			Level:   parser.LevelError,
			Message: "both tlsCertPath and tlsKeyPath must be set for TLS, or neither",
		})
	}

	if config.TLSCertPath != "" && config.TLSKeyPath != "" {
		if _, err := os.Stat(config.TLSCertPath); os.IsNotExist(err) {
			issues = append(issues, parser.ValidationIssue{
				Level:   parser.LevelWarning,
				Message: "tlsCertPath file does not exist",
			})
		}
		if _, err := os.Stat(config.TLSKeyPath); os.IsNotExist(err) {
			issues = append(issues, parser.ValidationIssue{
				Level:   parser.LevelWarning,
				Message: "tlsKeyPath file does not exist",
			})
		}
	}

	return issues
}

func init() {
	parser.RegisterServiceValidator(&TCPValidator{})
}
