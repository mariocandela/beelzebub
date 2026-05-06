package MCP

import (
	"testing"

	"github.com/beelzebub-labs/beelzebub/v3/internal/parser"
	"github.com/stretchr/testify/assert"
)

func TestMCPValidator_Name(t *testing.T) {
	v := &MCPValidator{}
	assert.Equal(t, "mcp", v.Name())
}

func TestMCPValidator_NotMCPProtocol(t *testing.T) {
	v := &MCPValidator{}
	config := parser.BeelzebubServiceConfiguration{
		Protocol: "http",
	}
	issues := v.Validate(config)
	assert.Empty(t, issues)
}

func TestMCPValidator_ToolWithParams(t *testing.T) {
	v := &MCPValidator{}
	config := parser.BeelzebubServiceConfiguration{
		Protocol: "mcp",
		Tools: []parser.Tool{
			{
				Name: "tool:user-account-manager",
				Params: []parser.Param{
					{Name: "user_id", Description: "The ID of the user"},
				},
			},
		},
	}
	issues := v.Validate(config)
	assert.Empty(t, issues)
}

func TestMCPValidator_ToolWithoutParams(t *testing.T) {
	v := &MCPValidator{}
	config := parser.BeelzebubServiceConfiguration{
		Protocol: "mcp",
		Tools: []parser.Tool{
			{
				Name:   "tool:no-params",
				Params: nil,
			},
		},
	}
	issues := v.Validate(config)
	assert.Len(t, issues, 1)
	assert.Equal(t, parser.LevelWarning, issues[0].Level)
	assert.Contains(t, issues[0].Message, "tool:no-params")
	assert.Contains(t, issues[0].Message, "has no parameters defined")
}

func TestMCPValidator_MultipleToolsOneEmpty(t *testing.T) {
	v := &MCPValidator{}
	config := parser.BeelzebubServiceConfiguration{
		Protocol: "mcp",
		Tools: []parser.Tool{
			{
				Name: "tool:with-params",
				Params: []parser.Param{
					{Name: "id", Description: "The ID"},
				},
			},
			{
				Name:   "tool:without-params",
				Params: []parser.Param{},
			},
		},
	}
	issues := v.Validate(config)
	assert.Len(t, issues, 1)
	assert.Equal(t, parser.LevelWarning, issues[0].Level)
	assert.Contains(t, issues[0].Message, "tool:without-params")
}

func TestMCPValidator_NoTools(t *testing.T) {
	v := &MCPValidator{}
	config := parser.BeelzebubServiceConfiguration{
		Protocol: "mcp",
		Tools:    []parser.Tool{},
	}
	issues := v.Validate(config)
	assert.Len(t, issues, 1)
	assert.Equal(t, parser.LevelWarning, issues[0].Level)
	assert.Equal(t, "MCP service has no tools defined", issues[0].Message)
}

func TestMCPValidator_ToolEmptyName(t *testing.T) {
	v := &MCPValidator{}
	config := parser.BeelzebubServiceConfiguration{
		Protocol: "mcp",
		Tools: []parser.Tool{
			{
				Name:   "",
				Params: []parser.Param{{Name: "id"}},
			},
		},
	}
	issues := v.Validate(config)
	assert.Len(t, issues, 1)
	assert.Equal(t, parser.LevelWarning, issues[0].Level)
	assert.Contains(t, issues[0].Message, "tool has no name defined")
}

func TestMCPValidator_WithTools(t *testing.T) {
	v := &MCPValidator{}
	config := parser.BeelzebubServiceConfiguration{
		Protocol: "mcp",
		Tools: []parser.Tool{
			{
				Name: "tool:user-account-manager",
				Params: []parser.Param{
					{Name: "user_id", Description: "The ID of the user"},
				},
			},
		},
	}
	issues := v.Validate(config)
	for _, issue := range issues {
		assert.NotEqual(t, "MCP service has no tools defined", issue.Message)
	}
}

func TestMCPValidator_ToolEmptyNameAndNoParams(t *testing.T) {
	v := &MCPValidator{}
	config := parser.BeelzebubServiceConfiguration{
		Protocol: "mcp",
		Tools: []parser.Tool{
			{
				Name:   "",
				Params: nil,
			},
		},
	}
	issues := v.Validate(config)
	assert.Len(t, issues, 2)
}
