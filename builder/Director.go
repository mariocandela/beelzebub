package builder

import (
	"beelzebub/parser"
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

	_, err := d.builder.buildLogger(beelzebubCoreConfigurations.Core.Logging)
	if err != nil {
		return nil, err
	}
	//TODO manage file.close() with method stopBeelzebub on this director
	//defer file.Close()

	if beelzebubCoreConfigurations.Core.Tracing.RabbitMQEnabled {
		//Manage close connection on rabbitMQ
		_, err := d.builder.buildRabbitMQ(beelzebubCoreConfigurations.Core.Tracing.RabbitMQURI)
		if err != nil {
			return nil, err
		}
		//TODO manage connection.Close() with method stopBeelzebub on this director
		//defer connection.Close()

	}

	//TODO Set tracing strategy

	return d.builder.build(), nil
}
