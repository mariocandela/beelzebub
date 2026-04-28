package cli

import (
	"fmt"

	"github.com/beelzebub-labs/beelzebub/v3/internal/parser"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	_ "github.com/beelzebub-labs/beelzebub/v3/internal/protocols/strategies/HTTP"
	_ "github.com/beelzebub-labs/beelzebub/v3/internal/protocols/strategies/MCP"
	_ "github.com/beelzebub-labs/beelzebub/v3/internal/protocols/strategies/SSH"
	_ "github.com/beelzebub-labs/beelzebub/v3/internal/protocols/strategies/TCP"
	_ "github.com/beelzebub-labs/beelzebub/v3/internal/protocols/strategies/TELNET"
)

var (
	validateConfCore     string
	validateConfServices string
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate configuration files without starting services",
	Long:  "Parse and validate core and service YAML configurations, reporting any errors and warnings.",
	RunE:  validateConfigurations,
}

func init() {
	validateCmd.Flags().StringVarP(&validateConfCore, "conf-core", "c", "./configurations/beelzebub.yaml", "Path to core configuration file")
	validateCmd.Flags().StringVarP(&validateConfServices, "conf-services", "s", "./configurations/services/", "Path to services configuration directory")
}

func validateConfigurations(_ *cobra.Command, _ []string) error {
	log.SetLevel(log.ErrorLevel)

	p := parser.Init(validateConfCore, validateConfServices)

	services, parseIssues, err := p.ReadConfigurationsServicesForValidation()
	if err != nil {
		return fmt.Errorf("services config: %w", err)
	}

	serviceResult := parser.Validate(services, parseIssues)

	var coreResult parser.ValidateResult
	coreConf, err := p.ReadConfigurationsCore()
	if err != nil {
		coreResult = parser.ValidateResult{
			Results: []parser.ValidationResult{
				{Filename: validateConfCore, Issues: []parser.ValidationIssue{
					{Level: parser.LevelError, Message: fmt.Sprintf("failed to read core config: %v", err)},
				}},
			},
			TotalErrors: 1,
		}
	} else {
		coreResult = parser.ValidateCore(coreConf, validateConfCore)
	}

	combined := parser.ValidateResult{
		Results:       append(serviceResult.Results, coreResult.Results...),
		TotalErrors:   serviceResult.TotalErrors + coreResult.TotalErrors,
		TotalWarnings: serviceResult.TotalWarnings + coreResult.TotalWarnings,
	}

	fmt.Printf("Validating services from %s\n", validateConfServices)
	fmt.Printf("Validating core from %s\n\n", validateConfCore)

	combined.Print()

	if combined.ExitCode() != 0 {
		return fmt.Errorf("validation failed with %d error(s)", combined.TotalErrors)
	}

	return nil
}
