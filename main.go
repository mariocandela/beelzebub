package main

import (
	"beelzebub/parser"
	"beelzebub/protocols"
	"beelzebub/tracer"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
)

var quit = make(chan struct{})

func main() {
	configurationsParser := parser.Init("./configurations/beelzebub.yaml", "./configurations/services/")

	coreConfigurations, err := configurationsParser.ReadConfigurationsCore()
	if err != nil {
		log.Fatal(err)
	}

	fileLogs := configureLogging(coreConfigurations.Core.Logging)
	defer fileLogs.Close()

	beelzebubServicesConfiguration, err := configurationsParser.ReadConfigurationsServices()
	if err != nil {
		log.Fatal(err)
	}

	// Protocol strategies
	secureShellStrategy := &protocols.SecureShellStrategy{}
	hypertextTransferProtocolStrategy := &protocols.HypertextTransferProtocolStrategy{}

	// Protocol manager
	serviceManager := protocols.InitProtocolManager(logStrategy, hypertextTransferProtocolStrategy)

	for _, beelzebubServiceConfiguration := range beelzebubServicesConfiguration {
		switch beelzebubServiceConfiguration.Protocol {
		case "http":
			serviceManager.SetProtocolStrategy(hypertextTransferProtocolStrategy)
			break
		case "ssh":
			serviceManager.SetProtocolStrategy(secureShellStrategy)
			break
		default:
			log.Fatalf("Protocol %s not managed", beelzebubServiceConfiguration.Protocol)
			continue
		}

		err := serviceManager.InitService(beelzebubServiceConfiguration)
		if err != nil {
			log.Errorf("Error during init protocol: %s, %s", beelzebubServiceConfiguration.Protocol, err.Error())
		}
	}
	<-quit
}
func logStrategy(event tracer.Event) {
	log.WithFields(log.Fields{
		"status": event.Status.String(),
		"event":  event,
	}).Info("New Event")
}

func configureLogging(configurations parser.Logging) *os.File {
	file, err := os.OpenFile(configurations.LogsPath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}

	log.SetOutput(io.MultiWriter(os.Stdout, file))

	log.SetFormatter(&log.JSONFormatter{
		DisableTimestamp: configurations.LogDisableTimestamp,
	})
	log.SetReportCaller(configurations.DebugReportCaller)
	if configurations.Debug {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}
	return file
}
