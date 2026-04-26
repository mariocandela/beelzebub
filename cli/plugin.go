package cli

import (
	"fmt"

	"github.com/mariocandela/beelzebub/v3/pkg/plugin"
	// Blank imports ensure built-in plugins self-register before the command runs.
	_ "github.com/mariocandela/beelzebub/v3/internal/plugins"
	"github.com/spf13/cobra"
)

var pluginCmd = &cobra.Command{
	Use:   "plugin",
	Short: "Manage and inspect plugins",
}

var pluginListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all registered plugins",
	Run:   listPlugins,
}

func init() {
	pluginCmd.AddCommand(pluginListCmd)
}

func listPlugins(_ *cobra.Command, _ []string) {
	metas := plugin.List()
	if len(metas) == 0 {
		fmt.Println("No plugins registered.")
		return
	}

	fmt.Printf("%-20s %-10s %-12s %s\n", "NAME", "VERSION", "AUTHOR", "DESCRIPTION")
	fmt.Printf("%-20s %-10s %-12s %s\n", "----", "-------", "------", "-----------")
	for _, m := range metas {
		fmt.Printf("%-20s %-10s %-12s %s\n", m.Name, m.Version, m.Author, m.Description)
	}
}
