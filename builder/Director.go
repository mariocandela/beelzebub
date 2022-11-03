package builder

import (
	"beelzebub/parser"
	"beelzebub/tracer"
	"context"
	"encoding/json"
	amqp "github.com/rabbitmq/amqp091-go"
	log "github.com/sirupsen/logrus"
)

type Director struct {
	builder *Builder
}

func NewDirector(builder *Builder) *Director {
	return &Director{
		builder: builder,
	}
}

func (d *Director) BuildBeelzebub(beelzebubCoreConfigurations *parser.BeelzebubCoreConfigurations, beelzebubServicesConfiguration []parser.BeelzebubServiceConfiguration) (*Builder, error) {
	d.builder.beelzebubServicesConfiguration = beelzebubServicesConfiguration

	if err := d.builder.buildLogger(beelzebubCoreConfigurations.Core.Logging); err != nil {
		return nil, err
	}

	d.builder.setTraceStrategy(d.standardOutStrategy)

	if beelzebubCoreConfigurations.Core.Tracing.RabbitMQEnabled {
		d.builder.setTraceStrategy(d.rabbitMQTraceStrategy)
		err := d.builder.buildRabbitMQ(beelzebubCoreConfigurations.Core.Tracing.RabbitMQURI)
		if err != nil {
			return nil, err
		}
	}

	//TODO Set tracing strategy

	return d.builder.build(), nil
}

func (d *Director) standardOutStrategy(event tracer.Event) {
	log.WithFields(log.Fields{
		"status": event.Status,
		"event":  event,
	}).Info("New Event")
}

func (d *Director) rabbitMQTraceStrategy(event tracer.Event) {
	log.WithFields(log.Fields{
		"status": event.Status,
		"event":  event,
	}).Info("New Event")

	log.Debug("Push Event on queue")
	eventJSON, err := json.Marshal(event)
	if err != nil {
		log.Error(err.Error())
		return
	}

	publishing := amqp.Publishing{ContentType: "application/json", Body: eventJSON}

	if err = d.builder.rabbitMQChannel.PublishWithContext(context.TODO(), "", RabbitmqQueueName, false, false, publishing); err != nil {
		log.Error(err.Error())
	}
}
