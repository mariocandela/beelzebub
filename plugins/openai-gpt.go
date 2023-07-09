package plugins

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	
	log "github.com/sirupsen/logrus"
	
	"github.com/go-resty/resty/v2"
)

const (
	// Reference: https://www.engraved.blog/building-a-virtual-machine-inside/
	promptVirtualizeLinuxTerminal = "I want you to act as a Linux terminal. I will type commands and you will reply with what the terminal should show. I want you to only reply with the terminal output inside one unique code block, and nothing else. Do no write explanations. Do not type commands unless I instruct you to do so.\n\nA:pwd\n\nQ:/home/user\n\n"
	ChatGPTPluginName = "OpenAIGPTLinuxTerminal"
	openAIGPTEndpoint = "https://api.openai.com/v1/completions"
) 

type History struct {
	Input, Output string
}
	
type OpenAIGPTVirtualTerminal struct {
	Histories               []History
	OpenAPIChatGPTSecretKey string
	client                  *resty.Client
}

func (openAIGPTVirtualTerminal *OpenAIGPTVirtualTerminal) InjectDependency() {
	if openAIGPTVirtualTerminal.client == nil {
		openAIGPTVirtualTerminal.client = resty.New()
	}
}

type Choice struct {
	Text         string      `json:"text"`
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
	Model            string   `json:"model"`
	Prompt           string   `json:"prompt"`
	Temperature      int      `json:"temperature"`
	MaxTokens        int      `json:"max_tokens"`
	TopP             int      `json:"top_p"`
	FrequencyPenalty int      `json:"frequency_penalty"`
	PresencePenalty  int      `json:"presence_penalty"`
	Stop             []string `json:"stop"`
}

func buildPrompt(histories []History, command string) string {
	var sb strings.Builder

	sb.WriteString(promptVirtualizeLinuxTerminal)

	for _, history := range histories {
		sb.WriteString(fmt.Sprintf("A:%s\n\nQ:%s\n\n", history.Input, history.Output))
	}
	// Append command to evaluate
	sb.WriteString(fmt.Sprintf("A:%s\n\nQ:", command))

	return sb.String()
}

func (openAIGPTVirtualTerminal *OpenAIGPTVirtualTerminal) GetCompletions(command string) (string, error) {
	requestJson, err := json.Marshal(gptRequest{
		Model:            "text-davinci-003",
		Prompt:           buildPrompt(openAIGPTVirtualTerminal.Histories, command),
		Temperature:      0,
		MaxTokens:        100,
		TopP:             1,
		FrequencyPenalty: 0,
		PresencePenalty:  0,
		Stop:             []string{"\n"},
	})
	if err != nil {
		return "", err
	}

	if openAIGPTVirtualTerminal.OpenAPIChatGPTSecretKey == "" {
		return "", errors.New("OpenAPIChatGPTSecretKey is empty")
	}

	response, err := openAIGPTVirtualTerminal.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(requestJson).
		SetAuthToken(openAIGPTVirtualTerminal.OpenAPIChatGPTSecretKey).
		SetResult(&gptResponse{}).
		Post(openAIGPTEndpoint)

	if err != nil {
		return "", err
	}
	log.Debug(response)
	if len(response.Result().(*gptResponse).Choices) == 0 {
		return "", errors.New("no choices")
	}

	return response.Result().(*gptResponse).Choices[0].Text, nil
}
