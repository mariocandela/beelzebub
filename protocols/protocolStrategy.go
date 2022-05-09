package protocols

import (
	"beelzebub/parser"
	"beelzebub/tracer"
)

type ServiceStrategy interface {
	Init(beelzebubServiceConfiguration parser.BeelzebubServiceConfiguration, tracer tracer.Tracer) error
}
