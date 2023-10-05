package main

import (
	"flag"
	"github.com/mariocandela/beelzebub/builder"
	"github.com/mariocandela/beelzebub/parser"

	log "github.com/sirupsen/logrus"
)

func main() {
	var (
		quit                            = make(chan struct{})
		configurationsCorePath          string
		configurationsServicesDirectory string
	)

	flag.StringVar(&configurationsCorePath, "confCore", "./configurations/beelzebub.yaml", "Provide the path of configurations core")
	flag.StringVar(&configurationsServicesDirectory, "confServices", "./configurations/services/", "Directory config services")
	flag.Parse()

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
