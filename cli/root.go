package cli

import "github.com/spf13/cobra"

const logo = `
██████  ███████ ███████ ██      ███████ ███████ ██████  ██    ██ ██████
██   ██ ██      ██      ██         ███  ██      ██   ██ ██    ██ ██   ██
██████  █████   █████   ██        ███   █████   ██████  ██    ██ ██████
██   ██ ██      ██      ██       ███    ██      ██   ██ ██    ██ ██   ██
██████  ███████ ███████ ███████ ███████ ███████ ██████   ██████  ██████
Honeypot Framework, happy hacking!
`

var rootCmd = &cobra.Command{
	Use:          "beelzebub",
	Short:        "A honeypot framework supporting SSH, HTTP, TCP, TELNET, and MCP",
	Long:         logo + "A honeypot framework supporting SSH, HTTP, TCP, TELNET, and MCP.",
	SilenceUsage: true,
}

// Execute is the entrypoint called by main.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(pluginCmd)
}
