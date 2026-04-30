package builder

import (
	"testing"

	"github.com/beelzebub-labs/beelzebub/v3/internal/parser"
	"github.com/beelzebub-labs/beelzebub/v3/internal/tracer"
)

func TestNewDirector(t *testing.T) {
	b := NewBuilder()
	d := NewDirector(b)

	if d.builder != b {
		t.Errorf("expected builder to be set")
	}
}

func TestBuildBeelzebub_Standard(t *testing.T) {
	b := NewBuilder()
	d := NewDirector(b)

	coreConfig := &parser.BeelzebubCoreConfigurations{}
	servicesConfig := []parser.BeelzebubServiceConfiguration{}

	result, err := d.BuildBeelzebub(coreConfig, servicesConfig)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result == nil {
		t.Fatalf("expected builder result, got nil")
	}

	if result.traceStrategy == nil {
		t.Errorf("expected traceStrategy to be set")
	}
}

func TestBuildBeelzebub_BeelzebubCloud(t *testing.T) {
	b := NewBuilder()
	d := NewDirector(b)

	coreConfig := &parser.BeelzebubCoreConfigurations{}
	coreConfig.Core.BeelzebubCloud.Enabled = true
	coreConfig.Core.BeelzebubCloud.URI = "http://localhost:8080"
	coreConfig.Core.BeelzebubCloud.AuthToken = "token"

	servicesConfig := []parser.BeelzebubServiceConfiguration{}

	result, err := d.BuildBeelzebub(coreConfig, servicesConfig)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.traceStrategy == nil {
		t.Errorf("expected traceStrategy to be set")
	}

	// Call strategy for coverage
	d.beelzebubCloudStrategy(tracer.Event{})
}

func TestStandardOutStrategy(t *testing.T) {
	d := NewDirector(NewBuilder())
	d.standardOutStrategy(tracer.Event{Protocol: "test"})
}

// Cannot easily test RabbitMQ connection in unit tests without a mock
