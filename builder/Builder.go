package builder

import (
	"beelzebub/parser"
	"beelzebub/tracer"
)

type Builder struct {
	beelzebubCoreConfigurations    parser.BeelzebubCoreConfigurations
	beelzebubServicesConfiguration []parser.BeelzebubServiceConfiguration
	traceStrategy                  tracer.Strategy
}

func (b Builder) setBeelzebubCoreConfigurations(beelzebubCoreConfigurations parser.BeelzebubCoreConfigurations) {
	//TODO implement me
	panic("implement me")
}

func (b Builder) setTraceStrategy(traceStrategy tracer.Strategy) {
	//TODO implement me
	panic("implement me")
}

func (b Builder) setLogger(configurations parser.Logging) {
	//TODO implement me
	panic("implement me")
}

func (b Builder) build() Beelzebub {
	//TODO implement me
	panic("implement me")
}

func newBuilder() *Builder {
	return &Builder{}
}
