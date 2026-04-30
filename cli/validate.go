package cli

import (
	"fmt"
	"strings"

	"github.com/beelzebub-labs/beelzebub/v3/internal/parser"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate configuration files without starting services",
	Long:  "Parse and validate core and service YAML configurations, reporting any errors.",
	RunE:  validateConfigurations,
}

func init() {
}

var knownProtocols = map[string]bool{
	"http": true, "ssh": true, "tcp": true, "telnet": true, "mcp": true,
}

func validateConfigurations(_ *cobra.Command, _ []string) error {
	// Let log level be controlled by rootLogLevel, don't force ErrorLevel here
	// unless user didn't specify, but root command handles setting default.

	p := parser.Init(rootConfCore, rootConfServices)

	coreConf, err := p.ReadConfigurationsCore()
	if err != nil {
		return fmt.Errorf("core config: %w", err)
	}

	printSection("Core configuration", rootConfCore)
	printField("Prometheus", formatOptional(coreConf.Core.Prometheus.Port+coreConf.Core.Prometheus.Path))
	printField("RabbitMQ", formatBool(coreConf.Core.Tracings.RabbitMQ.Enabled))
	printField("Beelzebub Cloud", formatBool(coreConf.Core.BeelzebubCloud.Enabled))

	services, err := p.ReadConfigurationsServices()
	if err != nil {
		return fmt.Errorf("services config: %w", err)
	}

	fmt.Println()
	printSection("Services", fmt.Sprintf("%s (%d found)", rootConfServices, len(services)))

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
