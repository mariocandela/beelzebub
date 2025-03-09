package plugins

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/mariocandela/beelzebub/v3/tracer"
	log "github.com/sirupsen/logrus"
	"os"
	"regexp"
	"strings"
)

const (
	systemPromptVirtualizeLinuxTerminal = "You will act as an Ubuntu Linux terminal. The user will type commands, and you are to reply with what the terminal should show. Your responses must be contained within a single code block. Do not provide note. Do not provide explanations or type commands unless explicitly instructed by the user. Your entire response/output is going to consist of a simple text with \n for new line, and you will NOT wrap it within string md markers"
	systemPromptVirtualizeHTTPServer    = "You will act as an unsecure HTTP Server with multiple vulnerability like aws and git credentials stored into root http directory. The user will send HTTP requests, and you are to reply with what the server should show. Do not provide explanations or type commands unless explicitly instructed by the user."
	LLMPluginName                       = "LLMHoneypot"
	openAIEndpoint                      = "https://api.openai.com/v1/chat/completions"
	ollamaEndpoint                      = "http://localhost:11434/api/chat"
)

type LLMHoneypot struct {
	Histories    []Message
	OpenAIKey    string
	client       *resty.Client
	Protocol     tracer.Protocol
	Provider     LLMProvider
	Model        string
	Host         string
	CustomPrompt string
}

type Choice struct {
	Message      Message `json:"message"`
	Index        int     `json:"index"`
	FinishReason string  `json:"finish_reason"`
}

type Response struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int      `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Message Message  `json:"message"`
	Usage   struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

type Request struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Role int

const (
	SYSTEM Role = iota
	USER
	ASSISTANT
)

func (role Role) String() string {
	return [...]string{"system", "user", "assistant"}[role]
}

type LLMProvider int

const (
	Ollama LLMProvider = iota
	OpenAI
)

func FromStringToLLMProvider(llmProvider string) (LLMProvider, error) {
	switch strings.ToLower(llmProvider) {
	case "ollama":
		return Ollama, nil
	case "openai":
		return OpenAI, nil
	default:
		return -1, fmt.Errorf("provider %s not found, valid providers: ollama, openai", llmProvider)
	}
}

func InitLLMHoneypot(config LLMHoneypot) *LLMHoneypot {
	// Inject the dependencies
	config.client = resty.New()

	if os.Getenv("OPEN_AI_SECRET_KEY") != "" {
		config.OpenAIKey = os.Getenv("OPEN_AI_SECRET_KEY")
	}

	return &config
}

func (llmHoneypot *LLMHoneypot) buildPrompt(command string) ([]Message, error) {
	var messages []Message
	var prompt string

	switch llmHoneypot.Protocol {
	case tracer.SSH:
		prompt = systemPromptVirtualizeLinuxTerminal
		if llmHoneypot.CustomPrompt != "" {
			prompt = llmHoneypot.CustomPrompt
		}
		messages = append(messages, Message{
			Role:    SYSTEM.String(),
			Content: prompt,
		})
		messages = append(messages, Message{
			Role:    USER.String(),
			Content: "pwd",
		})
		messages = append(messages, Message{
			Role:    ASSISTANT.String(),
			Content: "/home/user",
		})
		for _, history := range llmHoneypot.Histories {
			messages = append(messages, history)
		}
	case tracer.HTTP:
		prompt = systemPromptVirtualizeHTTPServer
		if llmHoneypot.CustomPrompt != "" {
			prompt = llmHoneypot.CustomPrompt
		}
		messages = append(messages, Message{
			Role:    SYSTEM.String(),
			Content: prompt,
		})
		messages = append(messages, Message{
			Role:    USER.String(),
			Content: "GET /index.html",
		})
		messages = append(messages, Message{
			Role:    ASSISTANT.String(),
			Content: "<html><body>Hello, World!</body></html>",
		})
	default:
		return nil, errors.New("no prompt for protocol selected")
	}
	messages = append(messages, Message{
		Role:    USER.String(),
		Content: command,
	})

	return messages, nil
}

func (llmHoneypot *LLMHoneypot) openAICaller(messages []Message) (string, error) {
	var err error

	requestJson, err := json.Marshal(Request{
		Model:    llmHoneypot.Model,
		Messages: messages,
		Stream:   false,
	})
	if err != nil {
		return "", err
	}

	if llmHoneypot.OpenAIKey == "" {
		return "", errors.New("openAIKey is empty")
	}

	if llmHoneypot.Host == "" {
		llmHoneypot.Host = openAIEndpoint
	}

	log.Debug(string(requestJson))
	response, err := llmHoneypot.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(requestJson).
		SetAuthToken(llmHoneypot.OpenAIKey).
		SetResult(&Response{}).
		Post(llmHoneypot.Host)

	if err != nil {
		return "", err
	}
	log.Debug(response)
	if len(response.Result().(*Response).Choices) == 0 {
		return "", errors.New("no choices")
	}

	return removeQuotes(response.Result().(*Response).Choices[0].Message.Content), nil
}

func (llmHoneypot *LLMHoneypot) ollamaCaller(messages []Message) (string, error) {
	var err error

	requestJson, err := json.Marshal(Request{
		Model:    llmHoneypot.Model,
		Messages: messages,
		Stream:   false,
	})
	if err != nil {
		return "", err
	}

	if llmHoneypot.Host == "" {
		llmHoneypot.Host = ollamaEndpoint
	}

	log.Debug(string(requestJson))
	response, err := llmHoneypot.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(requestJson).
		SetResult(&Response{}).
		Post(llmHoneypot.Host)

	if err != nil {
		return "", err
	}
	log.Debug(response)

	return removeQuotes(response.Result().(*Response).Message.Content), nil
}

func (llmHoneypot *LLMHoneypot) ExecuteModel(command string) (string, error) {
	var err error
	var prompt []Message

	prompt, err = llmHoneypot.buildPrompt(command)

	if err != nil {
		return "", err
	}

	switch llmHoneypot.Provider {
	case Ollama:
		return llmHoneypot.ollamaCaller(prompt)
	case OpenAI:
		return llmHoneypot.openAICaller(prompt)
	default:
		return "", fmt.Errorf("provider %d not found, valid providers: ollama, openai", llmHoneypot.Provider)
	}
}

func removeQuotes(content string) string {
	regex := regexp.MustCompile("(```( *)?([a-z]*)?(\\n)?)")
	return regex.ReplaceAllString(content, "")
}
