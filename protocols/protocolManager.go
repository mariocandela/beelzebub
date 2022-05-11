package protocols

import (
	"beelzebub/parser"
	"beelzebub/tracer"
)

type ServiceStrategy interface {
	Init(beelzebubServiceConfiguration parser.BeelzebubServiceConfiguration, tracer tracer.Tracer) error
}

type ProtocolManager struct {
	strategy ServiceStrategy
	tracer   *tracer.Tracer
}

func InitProtocolManager(tracerStrategy tracer.Strategy, strategy ServiceStrategy) *ProtocolManager {
	return &ProtocolManager{
		tracer:   tracer.Init(tracerStrategy),
		strategy: strategy,
	}
}

func (pm *ProtocolManager) SetProtocolStrategy(strategy ServiceStrategy) {
	pm.strategy = strategy
}

func (pm *ProtocolManager) InitService(beelzebubServiceConfiguration parser.BeelzebubServiceConfiguration) error {
	return pm.strategy.Init(beelzebubServiceConfiguration, *pm.tracer)
}
