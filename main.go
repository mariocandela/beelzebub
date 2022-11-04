package main

import (
	"beelzebub/builder"
	"beelzebub/parser"
	"fmt"
	log "github.com/sirupsen/logrus"
)

var quit = make(chan struct{})

func main() {
	parser := parser.Init("./configurations/beelzebub.yaml", "./configurations/services/")

	coreConfigurations, err := parser.ReadConfigurationsCore()
	failOnError(err, fmt.Sprintf("Error during ReadConfigurationsCore: "))

	beelzebubServicesConfiguration, err := parser.ReadConfigurationsServices()
	failOnError(err, fmt.Sprintf("Error during ReadConfigurationsServices: "))

	beelzebubBuilder := builder.NewBuilder()

	director := builder.NewDirector(beelzebubBuilder)

	beelzebubBuilder, err = director.BuildBeelzebub(coreConfigurations, beelzebubServicesConfiguration)
	if err != nil {
		log.Fatal(err)
	}

	if err := beelzebubBuilder.Run(); err != nil {
		log.Fatal(err)
		return
	}
	defer beelzebubBuilder.Close()

	<-quit
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
