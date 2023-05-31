package main

import (
	"beelzebub/builder"
	"beelzebub/parser"
	"flag"
	"fmt"

	log "github.com/sirupsen/logrus"
)

var quit = make(chan struct{})

func main() {
	var configurationsCorePath string
	var configurationsServicesDirectory string

	flag.StringVar(&configurationsCorePath, "confCore", "./configurations/beelzebub.yaml", "Provide the path of configurations core")
	flag.StringVar(&configurationsServicesDirectory, "confServices", "./configurations/services/", "Directory config services")
	flag.Parse()

	parser := parser.Init(configurationsCorePath, configurationsServicesDirectory)

	coreConfigurations, err := parser.ReadConfigurationsCore()
	failOnError(err, fmt.Sprintf("Error during ReadConfigurationsCore: "))

	beelzebubServicesConfiguration, err := parser.ReadConfigurationsServices()
	failOnError(err, fmt.Sprintf("Error during ReadConfigurationsServices: "))

	beelzebubBuilder := builder.NewBuilder()

	director := builder.NewDirector(beelzebubBuilder)

	beelzebubBuilder, err = director.BuildBeelzebub(coreConfigurations, beelzebubServicesConfiguration)
	failOnError(err, fmt.Sprintf("Error during BuildBeelzebub: "))

	err = beelzebubBuilder.Run()
	failOnError(err, fmt.Sprintf("Error during run beelzebub core: "))

	defer beelzebubBuilder.Close()

	<-quit
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
