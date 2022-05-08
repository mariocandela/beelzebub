package protocols

import "beelzebub/parser"

type ProtocolManager struct {
	strategy ServiceStrategy
}

func (pm *ProtocolManager) InitServiceManager() *ProtocolManager {
	return &ProtocolManager{}
}

func (pm *ProtocolManager) SetProtocolStrategy(strategy ServiceStrategy) {
	pm.strategy = strategy
}

func (pm *ProtocolManager) InitService(beelzebubServiceConfiguration parser.BeelzebubServiceConfiguration) error {
	return pm.strategy.Init(beelzebubServiceConfiguration)
}
