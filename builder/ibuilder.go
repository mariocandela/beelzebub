package builder

import (
	"beelzebub/parser"
	"beelzebub/tracer"
	"errors"
	"fmt"
)

type Beelzebub struct {
	beelzebubCoreConfigurations parser.BeelzebubCoreConfigurations
	traceStrategy               tracer.Strategy
	configurations              parser.Logging
	execute                     func() error
}

type IBuilder interface {
	setBeelzebubCoreConfigurations(beelzebubCoreConfigurations parser.BeelzebubCoreConfigurations)
	setTraceStrategy(traceStrategy tracer.Strategy)
	setLogger(configurations parser.Logging)
	build() Beelzebub
}

func getBuilder(builderType string) (IBuilder, error) {
	if builderType == "normal" {
		return newBuilder(), nil
	}

	return nil, errors.New(fmt.Sprintf("BuilderType %s not found", builderType))
}
