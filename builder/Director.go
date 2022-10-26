package builder

import (
	"beelzebub/parser"
)

type Director struct {
	builder Builder
}

func newDirector(builder Builder) *Director {
	return &Director{
		builder: builder,
	}
}

func (d *Director) setBuilder(builder Builder) {
	d.builder = builder
}

func (d *Director) buildBeelzebub(beelzebubCoreConfigurations *parser.BeelzebubCoreConfigurations) Builder {
	d.builder.buildLogger(beelzebubCoreConfigurations.Core.Logging)

	if beelzebubCoreConfigurations.Core.Tracing.RabbitMQEnabled {
		//Manage close connection on rabbitMQ
		d.builder.buildRabbitMQ(beelzebubCoreConfigurations.Core.Tracing.RabbitMQURI)

		//TODO Set tracing strategy
	}

	return d.builder.build()
}
