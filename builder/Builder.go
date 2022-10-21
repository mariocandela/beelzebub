package builder

import (
	"beelzebub/parser"
	"beelzebub/protocols"
	"beelzebub/tracer"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
)

type Beelzebub struct {
	beelzebubCoreConfigurations    parser.BeelzebubCoreConfigurations
	beelzebubServicesConfiguration []parser.BeelzebubServiceConfiguration
	traceStrategy                  tracer.Strategy
}

func (b *Beelzebub) setBeelzebubCoreConfigurations(beelzebubCoreConfigurations parser.BeelzebubCoreConfigurations) {
	b.beelzebubCoreConfigurations = beelzebubCoreConfigurations
}

func (b *Beelzebub) setTraceStrategy(traceStrategy tracer.Strategy) {
	b.traceStrategy = traceStrategy
}

func (b *Beelzebub) setLogger(configurations parser.Logging) {
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
}

func (b *Beelzebub) Run() error {
	// Init Protocol strategies
	secureShellStrategy := &protocols.SecureShellStrategy{}
	hypertextTransferProtocolStrategy := &protocols.HypertextTransferProtocolStrategy{}
	transmissionControlProtocolStrategy := &protocols.TransmissionControlProtocolStrategy{}

	// Init protocol manager, with simple log on stout trace strategy and default protocol HTTP
	protocolManager := protocols.InitProtocolManager(func(event tracer.Event) {
		log.WithFields(log.Fields{
			"status": event.Status,
			"event":  event,
		}).Info("New Event")
	}, hypertextTransferProtocolStrategy)

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

func (b *Beelzebub) build() *Beelzebub {
	return &Beelzebub{
		beelzebubCoreConfigurations:    b.beelzebubCoreConfigurations,
		beelzebubServicesConfiguration: b.beelzebubServicesConfiguration,
		traceStrategy:                  b.traceStrategy,
	}
}

func newBuilder() *Beelzebub {
	return &Beelzebub{}
}
