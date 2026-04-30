package cli

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const logo = `
██████  ███████ ███████ ██      ███████ ███████ ██████  ██    ██ ██████
██   ██ ██      ██      ██         ███  ██      ██   ██ ██    ██ ██   ██
██████  █████   █████   ██        ███   █████   ██████  ██    ██ ██████
██   ██ ██      ██      ██       ███    ██      ██   ██ ██    ██ ██   ██
██████  ███████ ███████ ███████ ███████ ███████ ██████   ██████  ██████
Deception runtime framework, happy hacking!
`

var (
	rootConfCore     string
	rootConfServices string
	rootLogLevel     string
)

var rootCmd = &cobra.Command{
	Use:          "beelzebub",
	Short:        "A Deception runtime framework supporting SSH, HTTP, TCP, TELNET, and MCP",
	Long:         logo + "A Deception runtime framework supporting SSH, HTTP, TCP, TELNET, and MCP.",
	SilenceUsage: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		level, err := log.ParseLevel(rootLogLevel)
		if err != nil {
			return err
		}
		log.SetLevel(level)
		return nil
	},
}

// Execute is the entrypoint called by main.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&rootConfCore, "conf-core", "c", "./configurations/beelzebub.yaml", "Path to core configuration file")
	rootCmd.PersistentFlags().StringVarP(&rootConfServices, "conf-services", "s", "./configurations/services/", "Path to services configuration directory")
	rootCmd.PersistentFlags().StringVarP(&rootLogLevel, "log-level", "l", "info", "Set log level (debug, info, warn, error, fatal, panic)")

	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(pluginCmd)
}
