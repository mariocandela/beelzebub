package HTTP

import (
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

	issues = append(issues, parser.ValidateTLSConfig(config.TLSCertPath, config.TLSKeyPath)...)

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
