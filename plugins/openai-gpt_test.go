package plugins

import (
	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
	"github.com/mariocandela/beelzebub/v3/tracer"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestBuildPromptEmptyHistory(t *testing.T) {
	//Given
	var histories []History
	command := "pwd"

	//When
	prompt := buildPrompt(histories, command)

	//Then
	assert.Equal(t,
		"You will act as an Ubuntu Linux terminal. The user will type commands, and you are to reply with what the terminal should show. Your responses must be contained within a single code block. Do not provide explanations or type commands unless explicitly instructed by the user. Remember previous commands and consider their effects on subsequent outputs.\n\nA:pwd\n\nQ:/home/user\n\nA:pwd\n\nQ:",
		prompt)
}

func TestBuildPromptWithHistory(t *testing.T) {
	//Given
	var histories = []History{
		{
			Input:  "cat hello.txt",
			Output: "world",
		},
		{
			Input:  "echo 1234",
			Output: "1234",
		},
	}

	command := "pwd"

	//When
	prompt := buildPrompt(histories, command)

	//Then
	assert.Equal(t,
		"You will act as an Ubuntu Linux terminal. The user will type commands, and you are to reply with what the terminal should show. Your responses must be contained within a single code block. Do not provide explanations or type commands unless explicitly instructed by the user. Remember previous commands and consider their effects on subsequent outputs.\n\nA:pwd\n\nQ:/home/user\n\nA:cat hello.txt\n\nQ:world\n\nA:echo 1234\n\nQ:1234\n\nA:pwd\n\nQ:",
		prompt)
}

func TestBuildGetCompletionsFailValidation(t *testing.T) {
	openAIGPTVirtualTerminal := Init(make([]History, 0), "", tracer.SSH)

	_, err := openAIGPTVirtualTerminal.GetCompletions("test")

	assert.Equal(t, "openAIKey is empty", err.Error())
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
						Text: "prova.txt",
					},
				},
			})
			if err != nil {
				return httpmock.NewStringResponse(500, ""), nil
			}
			return resp, nil
		},
	)

	openAIGPTVirtualTerminal := Init(make([]History, 0), "sdjdnklfjndslkjanfk", tracer.SSH)
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

	openAIGPTVirtualTerminal := Init(make([]History, 0), "sdjdnklfjndslkjanfk", tracer.SSH)
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
						Text: "[default]\nregion = us-west-2\noutput = json",
					},
				},
			})
			if err != nil {
				return httpmock.NewStringResponse(500, ""), nil
			}
			return resp, nil
		},
	)

	openAIGPTVirtualTerminal := Init(make([]History, 0), "sdjdnklfjndslkjanfk", tracer.HTTP)
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

	openAIGPTVirtualTerminal := Init(make([]History, 0), "sdjdnklfjndslkjanfk", tracer.HTTP)
	openAIGPTVirtualTerminal.client = client

	//When
	_, err := openAIGPTVirtualTerminal.GetCompletions("GET /.aws/credentials")

	//Then
	assert.Equal(t, "no choices", err.Error())
}
