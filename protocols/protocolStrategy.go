package protocols

import "beelzebub/parser"

type ServiceStrategy interface {
	Init(beelzebubServiceConfiguration parser.BeelzebubServiceConfiguration) error
}
