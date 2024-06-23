package plugins

import (
	"encoding/json"
	"errors"
	"github.com/go-resty/resty/v2"
	"github.com/mariocandela/beelzebub/v3/tracer"

	log "github.com/sirupsen/logrus"
)

const (
	systemPromptVirtualizeLinuxTerminal = "You will act as an Ubuntu Linux terminal. The user will type commands, and you are to reply with what the terminal should show. Your responses must be contained within a single code block. Do not provide explanations or type commands unless explicitly instructed by the user. Your entire response/output is going to consist of a simple text with \n for new line, and you will NOT wrap it within string md markers"
	systemPromptVirtualizeHTTPServer    = "You will act as an unsecure HTTP Server with multiple vulnerability like aws and git credentials stored into root http directory. The user will send HTTP requests, and you are to reply with what the server should show. Do not provide explanations or type commands unless explicitly instructed by the user."
	ChatGPTPluginName                   = "LLMHoneypot"
	openAIGPTEndpoint                   = "https://api.openai.com/v1/chat/completions"
)

type openAIVirtualHoneypot struct {
	Histories []Message
	openAIKey string
	client    *resty.Client
	protocol  tracer.Protocol
}

type Choice struct {
	Message      Message     `json:"message"`
	Index        int         `json:"index"`
	Logprobs     interface{} `json:"logprobs"`
	FinishReason string      `json:"finish_reason"`
}

type gptResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int      `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

type gptRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
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

func Init(history []Message, openAIKey string, protocol tracer.Protocol) *openAIVirtualHoneypot {
	return &openAIVirtualHoneypot{
		Histories: history,
		openAIKey: openAIKey,
		client:    resty.New(),
		protocol:  protocol,
	}
}

func buildPrompt(histories []Message, protocol tracer.Protocol, command string) ([]Message, error) {
	var messages []Message

	switch protocol {
	case tracer.SSH:
		messages = append(messages, Message{
			Role:    SYSTEM.String(),
			Content: systemPromptVirtualizeLinuxTerminal,
		})
		messages = append(messages, Message{
			Role:    USER.String(),
			Content: "pwd",
		})
		messages = append(messages, Message{
			Role:    ASSISTANT.String(),
			Content: "/home/user",
		})
		for _, history := range histories {
			messages = append(messages, history)
		}
	case tracer.HTTP:
		messages = append(messages, Message{
			Role:    SYSTEM.String(),
			Content: systemPromptVirtualizeHTTPServer,
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

func (openAIVirtualHoneypot *openAIVirtualHoneypot) GetCompletions(command string) (string, error) {
	var err error

	prompt, err := buildPrompt(openAIVirtualHoneypot.Histories, openAIVirtualHoneypot.protocol, command)

	if err != nil {
		return "", err
	}

	requestJson, err := json.Marshal(gptRequest{
		Model:    "gpt-4o",
		Messages: prompt,
	})
	if err != nil {
		return "", err
	}

	if openAIVirtualHoneypot.openAIKey == "" {
		return "", errors.New("openAIKey is empty")
	}

	log.Debug(string(requestJson))
	response, err := openAIVirtualHoneypot.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(requestJson).
		SetAuthToken(openAIVirtualHoneypot.openAIKey).
		SetResult(&gptResponse{}).
		Post(openAIGPTEndpoint)

	if err != nil {
		return "", err
	}
	log.Debug(response)
	if len(response.Result().(*gptResponse).Choices) == 0 {
		return "", errors.New("no choices")
	}

	return response.Result().(*gptResponse).Choices[0].Message.Content, nil
}
