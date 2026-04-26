package cli

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	"github.com/mariocandela/beelzebub/v3/internal/builder"
	"github.com/mariocandela/beelzebub/v3/internal/parser"
	"github.com/spf13/cobra"
)

var (
	runConfCore     string
	runConfServices string
	runMemLimitMiB  int
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Start the honeypot services",
	Long:  "Start all honeypot services defined in the configuration directory.",
	RunE:  runBeelzebub,
}

func init() {
	runCmd.Flags().StringVarP(&runConfCore, "conf-core", "c", "./configurations/beelzebub.yaml", "Path to core configuration file")
	runCmd.Flags().StringVarP(&runConfServices, "conf-services", "s", "./configurations/services/", "Path to services configuration directory")
	runCmd.Flags().IntVarP(&runMemLimitMiB, "mem-limit-mib", "m", 100, "Memory limit in MiB (-1 to disable)")
}

func runBeelzebub(cmd *cobra.Command, _ []string) error {
	if runMemLimitMiB > 0 {
		debug.SetMemoryLimit(int64(runMemLimitMiB) * 1024 * 1024)
	}

	p := parser.Init(runConfCore, runConfServices)

	coreConfigurations, err := p.ReadConfigurationsCore()
	if err != nil {
		return fmt.Errorf("reading core config: %w", err)
	}

	beelzebubServicesConfiguration, err := p.ReadConfigurationsServices()
	if err != nil {
		return fmt.Errorf("reading services config: %w", err)
	}

	if len(beelzebubServicesConfiguration) == 0 && !coreConfigurations.Core.BeelzebubCloud.Enabled {
		return errors.New("no services configured: provide a services directory, set BEELZEBUB_SERVICES_CONFIG, or enable beelzebub-cloud")
	}

	beelzebubBuilder := builder.NewBuilder()
	director := builder.NewDirector(beelzebubBuilder)

	beelzebubBuilder, err = director.BuildBeelzebub(coreConfigurations, beelzebubServicesConfiguration)
	if err != nil {
		return fmt.Errorf("building beelzebub: %w", err)
	}

	if err = beelzebubBuilder.Run(); err != nil {
		return fmt.Errorf("starting services: %w", err)
	}
	defer beelzebubBuilder.Close()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	fmt.Fprintf(cmd.OutOrStdout(), "\nReceived signal %s, shutting down...\n", sig)

	return nil
}
