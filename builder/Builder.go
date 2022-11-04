package builder

import (
	"beelzebub/parser"
	"beelzebub/protocols"
	"beelzebub/tracer"
	"errors"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
)

const RabbitmqQueueName = "event"

type Builder struct {
	beelzebubServicesConfiguration []parser.BeelzebubServiceConfiguration
	traceStrategy                  tracer.Strategy
	rabbitMQChannel                *amqp.Channel
	rabbitMQConnection             *amqp.Connection
	logsFile                       *os.File
}

func (b *Builder) setTraceStrategy(traceStrategy tracer.Strategy) {
	b.traceStrategy = traceStrategy
}

func (b *Builder) buildLogger(configurations parser.Logging) error {
	logsFile, err := os.OpenFile(configurations.LogsPath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return err
	}

	log.SetOutput(io.MultiWriter(os.Stdout, logsFile))

	log.SetFormatter(&log.JSONFormatter{
		DisableTimestamp: configurations.LogDisableTimestamp,
	})
	log.SetReportCaller(configurations.DebugReportCaller)
	if configurations.Debug {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}
	b.logsFile = logsFile
	return err
}

func (b *Builder) buildRabbitMQ(rabbitMQURI string) error {
	rabbitMQConnection, err := amqp.Dial(rabbitMQURI)
	if err != nil {
		return err
	}

	b.rabbitMQChannel, err = rabbitMQConnection.Channel()
	if err != nil {
		return err
	}

	//creates a queue if it doesn't already exist, or ensures that an existing queue matches the same parameters.
	if _, err = b.rabbitMQChannel.QueueDeclare(RabbitmqQueueName, false, false, false, false, nil); err != nil {
		return err
	}

	b.rabbitMQConnection = rabbitMQConnection
	return nil
}

func (b *Builder) Close() error {
	if err := b.rabbitMQChannel.Close(); err != nil {
		return err
	}
	if err := b.rabbitMQConnection.Close(); err != nil {
		return err
	}
	if err := b.logsFile.Close(); err != nil {
		return err
	}

	return nil
}

func (b *Builder) Run() error {
	// Init Protocol strategies
	secureShellStrategy := &protocols.SecureShellStrategy{}
	hypertextTransferProtocolStrategy := &protocols.HypertextTransferProtocolStrategy{}
	transmissionControlProtocolStrategy := &protocols.TransmissionControlProtocolStrategy{}

	// Init Tracer strategies, and set the trace strategy default HTTP
	protocolManager := protocols.InitProtocolManager(b.traceStrategy, hypertextTransferProtocolStrategy)

	for _, beelzebubServiceConfiguration := range b.beelzebubServicesConfiguration {
		switch beelzebubServiceConfiguration.Protocol {
		case "http":
			protocolManager.SetProtocolStrategy(hypertextTransferProtocolStrategy)
			break
		case "ssh":
			protocolManager.SetProtocolStrategy(secureShellStrategy)
			break
		case "tcp":
			protocolManager.SetProtocolStrategy(transmissionControlProtocolStrategy)
			break
		default:
			log.Fatalf("Protocol %s not managed", beelzebubServiceConfiguration.Protocol)
			continue
		}

		if err := protocolManager.InitService(beelzebubServiceConfiguration); err != nil {
			return errors.New(fmt.Sprintf("Error during init protocol: %s, %s", beelzebubServiceConfiguration.Protocol, err.Error()))
		}
	}
	return nil
}

func (b *Builder) build() *Builder {
	return &Builder{
		beelzebubServicesConfiguration: b.beelzebubServicesConfiguration,
		traceStrategy:                  b.traceStrategy,
	}
}

func NewBuilder() *Builder {
	return &Builder{}
}
