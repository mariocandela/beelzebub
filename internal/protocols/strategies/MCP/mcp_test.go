package MCP

import (
	"testing"

	"github.com/beelzebub-labs/beelzebub/v3/internal/parser"
	"github.com/beelzebub-labs/beelzebub/v3/internal/tracer"
	"github.com/stretchr/testify/assert"
)

type mockTracer struct {
	events []tracer.Event
}

func (m *mockTracer) TraceEvent(event tracer.Event) {
	m.events = append(m.events, event)
}

func TestMCPStrategy_Init_NoTools(t *testing.T) {
	strategy := &MCPStrategy{}
	mt := &mockTracer{}

	servConf := parser.BeelzebubServiceConfiguration{
		Address:     "127.0.0.1:0",
		Description: "test MCP server",
		Protocol:    "mcp",
		Tools:       []parser.Tool{},
	}

	err := strategy.Init(servConf, mt)
	assert.NoError(t, err)
}

func TestMCPStrategy_Init_ToolWithNoParams(t *testing.T) {
	strategy := &MCPStrategy{}
	mt := &mockTracer{}

	servConf := parser.BeelzebubServiceConfiguration{
		Address:     "127.0.0.1:0",
		Description: "test MCP server",
		Protocol:    "mcp",
		Tools: []parser.Tool{
			{
				Name:        "tool:no-params",
				Description: "A tool with no params",
				Params:      nil,
				Handler:     "response",
			},
		},
	}

	// Tool with no params should be skipped (logged as error) without panicking
	err := strategy.Init(servConf, mt)
	assert.NoError(t, err)
}

func TestMCPStrategy_Init_ToolWithParams(t *testing.T) {
	strategy := &MCPStrategy{}
	mt := &mockTracer{}

	readOnly := true
	destructive := false

	servConf := parser.BeelzebubServiceConfiguration{
		Address:     "127.0.0.1:0",
		Description: "test MCP server",
		Protocol:    "mcp",
		Tools: []parser.Tool{
			{
				Name:        "tool:query-logs",
				Description: "Query system logs",
				Annotations: &parser.ToolAnnotations{
					Title:           "Query Logs",
					ReadOnlyHint:    &readOnly,
					DestructiveHint: &destructive,
				},
				Params: []parser.Param{
					{Name: "filter", Description: "Log filter"},
				},
				Handler: "log_result",
			},
		},
	}

	err := strategy.Init(servConf, mt)
	assert.NoError(t, err)
}

func TestMCPStrategy_Init_ToolWithAllAnnotations(t *testing.T) {
	strategy := &MCPStrategy{}
	mt := &mockTracer{}

	trueVal := true
	falseVal := false

	servConf := parser.BeelzebubServiceConfiguration{
		Address:     "127.0.0.1:0",
		Description: "test MCP server",
		Protocol:    "mcp",
		Tools: []parser.Tool{
			{
				Name:        "tool:full-annotations",
				Description: "Tool with all annotations",
				Annotations: &parser.ToolAnnotations{
					Title:           "Full Annotations Tool",
					ReadOnlyHint:    &trueVal,
					DestructiveHint: &falseVal,
					IdempotentHint:  &trueVal,
					OpenWorldHint:   &falseVal,
				},
				Params: []parser.Param{
					{Name: "param1", Description: "First param"},
				},
				Handler: "handler_result",
			},
		},
	}

	err := strategy.Init(servConf, mt)
	assert.NoError(t, err)
}
