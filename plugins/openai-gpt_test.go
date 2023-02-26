package plugins

import (
	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
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
		"I want you to act as a Linux terminal. I will type commands and you will reply with what the terminal should show. I want you to only reply with the terminal output inside one unique code block, and nothing else. Do no write explanations. Do not type commands unless I instruct you to do so.\n\nA:pwd\n\nQ:/home/user\n\nA:pwd\n\nQ:",
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
		"I want you to act as a Linux terminal. I will type commands and you will reply with what the terminal should show. I want you to only reply with the terminal output inside one unique code block, and nothing else. Do no write explanations. Do not type commands unless I instruct you to do so.\n\nA:pwd\n\nQ:/home/user\n\nA:cat hello.txt\n\nQ:world\n\nA:echo 1234\n\nQ:1234\n\nA:pwd\n\nQ:",
		prompt)
}

func TestBuildGetCompletions(t *testing.T) {
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

	openAIGPTVirtualTerminal := OpenAIGPTVirtualTerminal{
		OpenAPIChatGPTSecretKey: "sdjdnklfjndslkjanfk",
		client:                  client,
	}

	//When
	str, err := openAIGPTVirtualTerminal.GetCompletions("ls")

	//Then
	assert.Nil(t, err)
	assert.Equal(t, "prova.txt", str)
}
