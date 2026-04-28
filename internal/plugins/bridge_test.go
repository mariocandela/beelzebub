package plugins

import (
	"testing"

	"github.com/beelzebub-labs/beelzebub/v3/internal/parser"
	"github.com/beelzebub-labs/beelzebub/v3/pkg/plugin"
	"github.com/stretchr/testify/assert"
)

func TestMessagesToPlugin(t *testing.T) {
	messages := []Message{
		{Role: "user", Content: "hello"},
		{Role: "assistant", Content: "world"},
	}

	result := MessagesToPlugin(messages)

	assert.Len(t, result, 2)
	assert.Equal(t, "user", result[0].Role)
	assert.Equal(t, "hello", result[0].Content)
	assert.Equal(t, "assistant", result[1].Role)
	assert.Equal(t, "world", result[1].Content)
}

func TestMessagesToPlugin_Empty(t *testing.T) {
	result := MessagesToPlugin([]Message{})
	assert.Empty(t, result)
}

func TestMessagesFromPlugin(t *testing.T) {
	messages := []plugin.Message{
		{Role: "user", Content: "ping"},
		{Role: "assistant", Content: "pong"},
	}

	result := MessagesFromPlugin(messages)

	assert.Len(t, result, 2)
	assert.Equal(t, "user", result[0].Role)
	assert.Equal(t, "ping", result[0].Content)
	assert.Equal(t, "assistant", result[1].Role)
	assert.Equal(t, "pong", result[1].Content)
}

func TestMessagesFromPlugin_Empty(t *testing.T) {
	result := MessagesFromPlugin([]plugin.Message{})
	assert.Empty(t, result)
}

func TestMessagesToPlugin_RoundTrip(t *testing.T) {
	original := []Message{
		{Role: "user", Content: "test content"},
	}

	pluginMsgs := MessagesToPlugin(original)
	restored := MessagesFromPlugin(pluginMsgs)

	assert.Equal(t, original, restored)
}

func TestConfigFromServiceConf(t *testing.T) {
	servConf := parser.BeelzebubServiceConfiguration{
		ServerVersion: "Apache/2.4.41",
		ServerName:    "test-server",
		Plugin: parser.Plugin{
			LLMProvider:             "openai",
			LLMModel:                "gpt-4",
			OpenAISecretKey:         "sk-test",
			Host:                    "localhost",
			Prompt:                  "custom prompt",
			InputValidationEnabled:  true,
			InputValidationPrompt:   "input check",
			OutputValidationEnabled: true,
			OutputValidationPrompt:  "output check",
			RateLimitEnabled:        true,
			RateLimitRequests:       10,
			RateLimitWindowSeconds:  60,
		},
	}

	cfg := ConfigFromServiceConf(servConf)

	assert.Equal(t, "openai", cfg.LLMProvider)
	assert.Equal(t, "gpt-4", cfg.LLMModel)
	assert.Equal(t, "sk-test", cfg.OpenAISecretKey)
	assert.Equal(t, "localhost", cfg.Host)
	assert.Equal(t, "custom prompt", cfg.Prompt)
	assert.True(t, cfg.InputValidationEnabled)
	assert.Equal(t, "input check", cfg.InputValidationPrompt)
	assert.True(t, cfg.OutputValidationEnabled)
	assert.Equal(t, "output check", cfg.OutputValidationPrompt)
	assert.True(t, cfg.RateLimitEnabled)
	assert.Equal(t, 10, cfg.RateLimitRequests)
	assert.Equal(t, 60, cfg.RateLimitWindowSeconds)
	assert.Equal(t, "Apache/2.4.41", cfg.ServerVersion)
	assert.Equal(t, "test-server", cfg.ServerName)
}

func TestConfigFromServiceConf_Defaults(t *testing.T) {
	cfg := ConfigFromServiceConf(parser.BeelzebubServiceConfiguration{})

	assert.Equal(t, "", cfg.LLMProvider)
	assert.Equal(t, "", cfg.LLMModel)
	assert.False(t, cfg.InputValidationEnabled)
	assert.False(t, cfg.OutputValidationEnabled)
	assert.False(t, cfg.RateLimitEnabled)
}
