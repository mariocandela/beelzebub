package HTTP

import (
	"os"

	"github.com/beelzebub-labs/beelzebub/v3/internal/parser"
)

type HTTPValidator struct{}

func (v *HTTPValidator) Name() string {
	return "http"
}

func (v *HTTPValidator) Validate(config parser.BeelzebubServiceConfiguration) []parser.ValidationIssue {
	if config.Protocol != "http" {
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

	if len(config.Commands) > 0 && config.FallbackCommand.Handler == "" && config.FallbackCommand.Plugin == "" {
		issues = append(issues, parser.ValidationIssue{
			Level:   parser.LevelWarning,
			Message: "HTTP service has commands but no fallbackCommand — unmatched requests will return empty 200 OK",
		})
	}

	return issues
}

func init() {
	parser.RegisterServiceValidator(&HTTPValidator{})
}
