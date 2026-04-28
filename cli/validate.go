package cli

import (
	"fmt"
	"strings"

	"github.com/beelzebub-labs/beelzebub/v3/internal/parser"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	validateConfCore     string
	validateConfServices string
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate configuration files without starting services",
	Long:  "Parse and validate core and service YAML configurations, reporting any errors.",
	RunE:  validateConfigurations,
}

func init() {
	validateCmd.Flags().StringVarP(&validateConfCore, "conf-core", "c", "./configurations/beelzebub.yaml", "Path to core configuration file")
	validateCmd.Flags().StringVarP(&validateConfServices, "conf-services", "s", "./configurations/services/", "Path to services configuration directory")
}

var knownProtocols = map[string]bool{
	"http": true, "ssh": true, "tcp": true, "telnet": true, "mcp": true,
}

func validateConfigurations(_ *cobra.Command, _ []string) error {
	// suppress logrus noise during validation
	log.SetLevel(log.ErrorLevel)

	p := parser.Init(validateConfCore, validateConfServices)

	coreConf, err := p.ReadConfigurationsCore()
	if err != nil {
		return fmt.Errorf("core config: %w", err)
	}

	printSection("Core configuration", validateConfCore)
	printField("Prometheus", formatOptional(coreConf.Core.Prometheus.Port+coreConf.Core.Prometheus.Path))
	printField("RabbitMQ", formatBool(coreConf.Core.Tracings.RabbitMQ.Enabled))
	printField("Beelzebub Cloud", formatBool(coreConf.Core.BeelzebubCloud.Enabled))

	services, err := p.ReadConfigurationsServices()
	if err != nil {
		return fmt.Errorf("services config: %w", err)
	}

	fmt.Println()
	printSection("Services", fmt.Sprintf("%s (%d found)", validateConfServices, len(services)))

	for i, svc := range services {
		if !knownProtocols[svc.Protocol] {
			return fmt.Errorf("service[%d] %q: unknown protocol %q", i+1, svc.Address, svc.Protocol)
		}

		extras := []string{}
		if svc.Plugin.LLMProvider != "" {
			extras = append(extras, fmt.Sprintf("plugin:%s/%s", svc.Plugin.LLMProvider, svc.Plugin.LLMModel))
		}
		if svc.Plugin.RateLimitEnabled {
			extras = append(extras, "rate-limited")
		}
		suffix := ""
		if len(extras) > 0 {
			suffix = "  [" + strings.Join(extras, ", ") + "]"
		}
		desc := svc.Description
		if desc == "" {
			desc = svc.ServerName
		}
		fmt.Printf("  [%d] %-7s %-22s %s%s\n", i+1, svc.Protocol, svc.Address, desc, suffix)
	}

	fmt.Println("\nAll configurations are valid.")
	return nil
}

func printSection(title, detail string) {
	fmt.Printf("%s: %s\n", title, detail)
}

func printField(name, value string) {
	fmt.Printf("  %-18s %s\n", name+":", value)
}

func formatBool(v bool) string {
	if v {
		return "enabled"
	}
	return "disabled"
}

func formatOptional(s string) string {
	if s == "" {
		return "(not set)"
	}
	return s
}
