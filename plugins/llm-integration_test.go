package plugins

import (
	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
	"github.com/mariocandela/beelzebub/v3/tracer"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

const SystemPromptLen = 4

func TestBuildPromptEmptyHistory(t *testing.T) {
	//Given
	var histories []Message
	command := "pwd"

	//When
	prompt, err := buildPrompt(histories, tracer.SSH, command)

	//Then
	assert.Nil(t, err)
	assert.Equal(t, SystemPromptLen, len(prompt))
}

func TestBuildPromptWithHistory(t *testing.T) {
	//Given
	var histories = []Message{
		{
			Role:    "cat hello.txt",
			Content: "world",
		},
	}

	command := "pwd"

	//When
	prompt, err := buildPrompt(histories, tracer.SSH, command)

	//Then
	assert.Nil(t, err)
	assert.Equal(t, SystemPromptLen+1, len(prompt))
}

func TestBuildGetCompletionsFailValidation(t *testing.T) {

	llmHoneypot := LLMHoneypot{
		Histories: make([]Message, 0),
		OpenAIKey: "",
		Protocol:  tracer.SSH,
		Model:     GPT4O,
	}

	openAIGPTVirtualTerminal := InitLLMHoneypot(llmHoneypot)

	_, err := openAIGPTVirtualTerminal.ExecuteModel("test")

	assert.Equal(t, "openAIKey is empty", err.Error())
}

func TestBuildGetCompletionsFailValidationStrategyType(t *testing.T) {

	llmHoneypot := LLMHoneypot{
		Histories: make([]Message, 0),
		OpenAIKey: "",
		Protocol:  tracer.TCP,
		Model:     GPT4O,
	}

	openAIGPTVirtualTerminal := InitLLMHoneypot(llmHoneypot)

	_, err := openAIGPTVirtualTerminal.ExecuteModel("test")

	assert.Equal(t, "no prompt for protocol selected", err.Error())
}

func TestBuildGetCompletionsSSHWithResultsOpenAI(t *testing.T) {
	client := resty.New()
	httpmock.ActivateNonDefault(client.GetClient())
	defer httpmock.DeactivateAndReset()

	// Given
	httpmock.RegisterResponder("POST", openAIGPTEndpoint,
		func(req *http.Request) (*http.Response, error) {
			resp, err := httpmock.NewJsonResponse(200, &Response{
				Choices: []Choice{
					{
						Message: Message{
							Role:    SYSTEM.String(),
							Content: "prova.txt",
						},
					},
				},
			})
			if err != nil {
				return httpmock.NewStringResponse(500, ""), nil
			}
			return resp, nil
		},
	)

	llmHoneypot := LLMHoneypot{
		Histories: make([]Message, 0),
		OpenAIKey: "sdjdnklfjndslkjanfk",
		Protocol:  tracer.SSH,
		Model:     GPT4O,
	}

	openAIGPTVirtualTerminal := InitLLMHoneypot(llmHoneypot)
	openAIGPTVirtualTerminal.client = client

	//When
	str, err := openAIGPTVirtualTerminal.ExecuteModel("ls")

	//Then
	assert.Nil(t, err)
	assert.Equal(t, "prova.txt", str)
}

func TestBuildGetCompletionsSSHWithResultsLLama(t *testing.T) {
	client := resty.New()
	httpmock.ActivateNonDefault(client.GetClient())
	defer httpmock.DeactivateAndReset()

	// Given
	httpmock.RegisterResponder("POST", ollamaEndpoint,
		func(req *http.Request) (*http.Response, error) {
			resp, err := httpmock.NewJsonResponse(200, &Response{
				Message: Message{
					Role:    SYSTEM.String(),
					Content: "prova.txt",
				},
			})
			if err != nil {
				return httpmock.NewStringResponse(500, ""), nil
			}
			return resp, nil
		},
	)

	llmHoneypot := LLMHoneypot{
		Histories: make([]Message, 0),
		Protocol:  tracer.SSH,
		Model:     LLAMA3,
	}

	openAIGPTVirtualTerminal := InitLLMHoneypot(llmHoneypot)
	openAIGPTVirtualTerminal.client = client

	//When
	str, err := openAIGPTVirtualTerminal.ExecuteModel("ls")

	//Then
	assert.Nil(t, err)
	assert.Equal(t, "prova.txt", str)
}

func TestBuildGetCompletionsSSHWithoutResults(t *testing.T) {
	client := resty.New()
	httpmock.ActivateNonDefault(client.GetClient())
	defer httpmock.DeactivateAndReset()

	// Given
	httpmock.RegisterResponder("POST", openAIGPTEndpoint,
		func(req *http.Request) (*http.Response, error) {
			resp, err := httpmock.NewJsonResponse(200, &Response{
				Choices: []Choice{},
			})
			if err != nil {
				return httpmock.NewStringResponse(500, ""), nil
			}
			return resp, nil
		},
	)

	llmHoneypot := LLMHoneypot{
		Histories: make([]Message, 0),
		OpenAIKey: "sdjdnklfjndslkjanfk",
		Protocol:  tracer.SSH,
		Model:     GPT4O,
	}

	openAIGPTVirtualTerminal := InitLLMHoneypot(llmHoneypot)
	openAIGPTVirtualTerminal.client = client

	//When
	_, err := openAIGPTVirtualTerminal.ExecuteModel("ls")

	//Then
	assert.Equal(t, "no choices", err.Error())
}

func TestBuildGetCompletionsHTTPWithResults(t *testing.T) {
	client := resty.New()
	httpmock.ActivateNonDefault(client.GetClient())
	defer httpmock.DeactivateAndReset()

	// Given
	httpmock.RegisterResponder("POST", openAIGPTEndpoint,
		func(req *http.Request) (*http.Response, error) {
			resp, err := httpmock.NewJsonResponse(200, &Response{
				Choices: []Choice{
					{
						Message: Message{
							Role:    SYSTEM.String(),
							Content: "[default]\nregion = us-west-2\noutput = json",
						},
					},
				},
			})
			if err != nil {
				return httpmock.NewStringResponse(500, ""), nil
			}
			return resp, nil
		},
	)

	llmHoneypot := LLMHoneypot{
		Histories: make([]Message, 0),
		OpenAIKey: "sdjdnklfjndslkjanfk",
		Protocol:  tracer.HTTP,
		Model:     GPT4O,
	}

	openAIGPTVirtualTerminal := InitLLMHoneypot(llmHoneypot)
	openAIGPTVirtualTerminal.client = client

	//When
	str, err := openAIGPTVirtualTerminal.ExecuteModel("GET /.aws/credentials")

	//Then
	assert.Nil(t, err)
	assert.Equal(t, "[default]\nregion = us-west-2\noutput = json", str)
}

func TestBuildGetCompletionsHTTPWithoutResults(t *testing.T) {
	client := resty.New()
	httpmock.ActivateNonDefault(client.GetClient())
	defer httpmock.DeactivateAndReset()

	// Given
	httpmock.RegisterResponder("POST", openAIGPTEndpoint,
		func(req *http.Request) (*http.Response, error) {
			resp, err := httpmock.NewJsonResponse(200, &Response{
				Choices: []Choice{},
			})
			if err != nil {
				return httpmock.NewStringResponse(500, ""), nil
			}
			return resp, nil
		},
	)

	llmHoneypot := LLMHoneypot{
		Histories: make([]Message, 0),
		OpenAIKey: "sdjdnklfjndslkjanfk",
		Protocol:  tracer.HTTP,
		Model:     GPT4O,
	}

	openAIGPTVirtualTerminal := InitLLMHoneypot(llmHoneypot)
	openAIGPTVirtualTerminal.client = client

	//When
	_, err := openAIGPTVirtualTerminal.ExecuteModel("GET /.aws/credentials")

	//Then
	assert.Equal(t, "no choices", err.Error())
}
