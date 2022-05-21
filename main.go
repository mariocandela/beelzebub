package main

import (
	"beelzebub/parser"
	"beelzebub/protocols"
	"beelzebub/tracer"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"io"
	"os"
)

var quit = make(chan struct{})

var channel *amqp.Channel

func main() {
	parser := parser.Init("./configurations/beelzebub.yaml", "./configurations/services/")

	coreConfigurations, err := parser.ReadConfigurationsCore()
	failOnError(err, fmt.Sprintf("Error during coreConfigurations: "))

	fileLogs := configureLoggingByConfigurations(coreConfigurations.Core.Logging)
	defer fileLogs.Close()

	beelzebubServicesConfiguration, err := parser.ReadConfigurationsServices()
	failOnError(err, fmt.Sprintf("Error during ReadConfigurationsServices: "))

	if coreConfigurations.Core.Tracing.RabbitMQEnabled {
		conn, err := amqp.Dial(coreConfigurations.Core.Tracing.RabbitMQURI)
		failOnError(err, "Failed to connect to RabbitMQ")
		defer conn.Close()

		channel, err := conn.Channel()
		failOnError(err, "Failed to open a channel")
		defer channel.Close()
	}

	// Init Protocol strategies
	secureShellStrategy := &protocols.SecureShellStrategy{}
	hypertextTransferProtocolStrategy := &protocols.HypertextTransferProtocolStrategy{}

	// Init protocol manager, with simple log on stout trace strategy and default protocol HTTP
	protocolManager := protocols.InitProtocolManager(traceStrategyStdout, hypertextTransferProtocolStrategy)

	for _, beelzebubServiceConfiguration := range beelzebubServicesConfiguration {
		switch beelzebubServiceConfiguration.Protocol {
		case "http":
			protocolManager.SetProtocolStrategy(hypertextTransferProtocolStrategy)
			break
		case "ssh":
			protocolManager.SetProtocolStrategy(secureShellStrategy)
			break
		default:
			log.Fatalf("Protocol %s not managed", beelzebubServiceConfiguration.Protocol)
			continue
		}

		err := protocolManager.InitService(beelzebubServiceConfiguration)
		failOnError(err, fmt.Sprintf("Error during init protocol: %s, ", beelzebubServiceConfiguration.Protocol))
	}
	<-quit
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func traceStrategyStdout(event tracer.Event) {
	log.WithFields(log.Fields{
		"status": event.Status,
		"event":  event,
	}).Info("New Event")

	if channel != nil {
		eventJSON, err := json.Marshal(event)
		failOnError(err, "Failed to Marshal Event")

		queue, err := channel.QueueDeclare(
			"event",
			false,
			false,
			false,
			false,
			nil,
		)
		failOnError(err, "Failed to declare a queue")

		err = channel.Publish(
			"",
			queue.Name,
			false,
			false,
			amqp.Publishing{
				ContentType: "application/json",
				Body:        eventJSON,
			})
		failOnError(err, "Failed to publish a message")
	}
}

func configureLoggingByConfigurations(configurations parser.Logging) *os.File {
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
