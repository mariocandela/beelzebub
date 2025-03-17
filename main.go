package main

import (
	"flag"
	"runtime/debug"

	"github.com/mariocandela/beelzebub/v3/builder"
	"github.com/mariocandela/beelzebub/v3/parser"

	log "github.com/sirupsen/logrus"
)

func main() {
	var (
		quit                            = make(chan struct{})
		configurationsCorePath          string
		configurationsServicesDirectory string
		memLimitMiB                     int
	)

	flag.StringVar(&configurationsCorePath, "confCore", "./configurations/beelzebub.yaml", "Provide the path of configurations core")
	flag.StringVar(&configurationsServicesDirectory, "confServices", "./configurations/services/", "Directory config services")
	flag.IntVar(&memLimitMiB, "memLimitMiB", 100, "Process Memory in MiB (default 100, set to -1 to use system default)")
	flag.Parse()

	if memLimitMiB > 0 {
		// SetMemoryLimit takes an int64 value for the number of bytes.
		// bytes value = MiB value * 1024 * 1024
		debug.SetMemoryLimit(int64(memLimitMiB * 1024 * 1024))
	}

	parser := parser.Init(configurationsCorePath, configurationsServicesDirectory)

	coreConfigurations, err := parser.ReadConfigurationsCore()
	failOnError(err, "Error during ReadConfigurationsCore: ")

	beelzebubServicesConfiguration, err := parser.ReadConfigurationsServices()
	failOnError(err, "Error during ReadConfigurationsServices: ")

	beelzebubBuilder := builder.NewBuilder()

	director := builder.NewDirector(beelzebubBuilder)

	beelzebubBuilder, err = director.BuildBeelzebub(coreConfigurations, beelzebubServicesConfiguration)
	failOnError(err, "Error during BuildBeelzebub: ")

	err = beelzebubBuilder.Run()
	failOnError(err, "Error during run beelzebub core: ")

	defer beelzebubBuilder.Close()

	<-quit
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
