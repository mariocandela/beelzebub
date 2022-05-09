package protocols

import (
	"beelzebub/parser"
	"beelzebub/tracer"
)

type ProtocolManager struct {
	strategy ServiceStrategy
	tracer   *tracer.Tracer
}

func (pm *ProtocolManager) InitServiceManager() *ProtocolManager {
	return &ProtocolManager{
		tracer: tracer.Init(),
	}
}

func (pm *ProtocolManager) SetProtocolStrategy(strategy ServiceStrategy) {
	pm.strategy = strategy
}

func (pm *ProtocolManager) InitService(beelzebubServiceConfiguration parser.BeelzebubServiceConfiguration) error {
	return pm.strategy.Init(beelzebubServiceConfiguration, *pm.tracer)
}
