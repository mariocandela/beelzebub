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
	openAIGPTVirtualTerminal := Init(make([]Message, 0), "", tracer.SSH)

	_, err := openAIGPTVirtualTerminal.GetCompletions("test")

	assert.Equal(t, "openAIKey is empty", err.Error())
}

func TestBuildGetCompletionsFailValidationStrategyType(t *testing.T) {
	openAIGPTVirtualTerminal := Init(make([]Message, 0), "", tracer.TCP)

	_, err := openAIGPTVirtualTerminal.GetCompletions("test")

	assert.Equal(t, "no prompt for protocol selected", err.Error())
}

func TestBuildGetCompletionsSSHWithResults(t *testing.T) {
	client := resty.New()
	httpmock.ActivateNonDefault(client.GetClient())
	defer httpmock.DeactivateAndReset()

	// Given
	httpmock.RegisterResponder("POST", openAIGPTEndpoint,
		func(req *http.Request) (*http.Response, error) {
			resp, err := httpmock.NewJsonResponse(200, &gptResponse{
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

	openAIGPTVirtualTerminal := Init(make([]Message, 0), "sdjdnklfjndslkjanfk", tracer.SSH)
	openAIGPTVirtualTerminal.client = client

	//When
	str, err := openAIGPTVirtualTerminal.GetCompletions("ls")

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
			resp, err := httpmock.NewJsonResponse(200, &gptResponse{
				Choices: []Choice{},
			})
			if err != nil {
				return httpmock.NewStringResponse(500, ""), nil
			}
			return resp, nil
		},
	)

	openAIGPTVirtualTerminal := Init(make([]Message, 0), "sdjdnklfjndslkjanfk", tracer.SSH)
	openAIGPTVirtualTerminal.client = client

	//When
	_, err := openAIGPTVirtualTerminal.GetCompletions("ls")

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
			resp, err := httpmock.NewJsonResponse(200, &gptResponse{
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

	openAIGPTVirtualTerminal := Init(make([]Message, 0), "sdjdnklfjndslkjanfk", tracer.HTTP)
	openAIGPTVirtualTerminal.client = client

	//When
	str, err := openAIGPTVirtualTerminal.GetCompletions("GET /.aws/credentials")

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
			resp, err := httpmock.NewJsonResponse(200, &gptResponse{
				Choices: []Choice{},
			})
			if err != nil {
				return httpmock.NewStringResponse(500, ""), nil
			}
			return resp, nil
		},
	)

	openAIGPTVirtualTerminal := Init(make([]Message, 0), "sdjdnklfjndslkjanfk", tracer.HTTP)
	openAIGPTVirtualTerminal.client = client

	//When
	_, err := openAIGPTVirtualTerminal.GetCompletions("GET /.aws/credentials")

	//Then
	assert.Equal(t, "no choices", err.Error())
}
