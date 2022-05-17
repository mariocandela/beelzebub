package main

import (
	"beelzebub/parser"
	"beelzebub/protocols"
	"beelzebub/tracer"
	"context"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"io"
	"os"
	"time"
)

var quit = make(chan struct{})

var mongoClient *mongo.Client

func main() {
	parser := parser.Init("./configurations/beelzebub.yaml", "./configurations/services/")

	coreConfigurations, err := parser.ReadConfigurationsCore()
	if err != nil {
		log.Fatal(err)
	}

	fileLogs := configureLoggingByConfigurations(coreConfigurations.Core.Logging)
	defer fileLogs.Close()

	beelzebubServicesConfiguration, err := parser.ReadConfigurationsServices()
	if err != nil {
		log.Fatal(err)
	}

	// Init Protocol strategies
	secureShellStrategy := &protocols.SecureShellStrategy{}
	hypertextTransferProtocolStrategy := &protocols.HypertextTransferProtocolStrategy{}

	// Init protocol manager, with simple log on stout trace strategy and default protocol HTTP
	protocolManager := protocols.InitProtocolManager(traceStrategyStdout, hypertextTransferProtocolStrategy)

	if coreConfigurations.Core.Tracing.MongoEnabled {
		mongoClient = buildMongoClient(coreConfigurations.Core.Tracing.MongoURI)
		defer func(mongoClient *mongo.Client, ctx context.Context) {
			err := mongoClient.Disconnect(ctx)
			if err != nil {
				log.Error(err)
			}
		}(mongoClient, context.TODO())
	}

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
		if err != nil {
			log.Errorf("Error during init protocol: %s, %s", beelzebubServiceConfiguration.Protocol, err.Error())
		}
	}
	<-quit
}
func traceStrategyStdout(event tracer.Event) {
	log.WithFields(log.Fields{
		"status": event.Status,
		"event":  event,
	}).Info("New Event")

	if mongoClient != nil {
		coll := mongoClient.Database("beelzebub").Collection("event")
		data, err := bson.Marshal(event)
		if err != nil {
			log.Fatal(err)
		}

		_, err = coll.InsertOne(context.TODO(), data)
		if err != nil {
			log.Fatal(err)
		}
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

func buildMongoClient(uri string) *mongo.Client {
	// Create a new client and connect to the server
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri).SetServerSelectionTimeout(time.Second*2))
	if err != nil {
		log.Fatal(err)
	}
	// Ping the primary
	if err := client.Ping(context.TODO(), readpref.Primary()); err != nil {
		log.Fatal(err)
	}
	log.Println("Successfully connected and pinged.")
	return client
}
