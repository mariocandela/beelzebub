package cli

import (
	"fmt"
	"runtime"
	"runtime/debug"

	"github.com/spf13/cobra"
)

// These vars are overridden at build time via -ldflags.
var (
	Version   = "dev"
	CommitSHA = "unknown"
	BuildDate = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run:   printVersion,
}

func printVersion(_ *cobra.Command, _ []string) {
	version := Version
	commit := CommitSHA

	// Fall back to module build info when not set via ldflags.
	if version == "dev" {
		if info, ok := debug.ReadBuildInfo(); ok {
			if info.Main.Version != "" && info.Main.Version != "(devel)" {
				version = info.Main.Version
			}
			for _, s := range info.Settings {
				if s.Key == "vcs.revision" && commit == "unknown" {
					if len(s.Value) > 8 {
						commit = s.Value[:8]
					} else {
						commit = s.Value
					}
				}
			}
		}
	}

	fmt.Printf("beelzebub %s\n", version)
	fmt.Printf("  commit:     %s\n", commit)
	fmt.Printf("  build date: %s\n", BuildDate)
	fmt.Printf("  go:         %s\n", runtime.Version())
	fmt.Printf("  os/arch:    %s/%s\n", runtime.GOOS, runtime.GOARCH)
}
