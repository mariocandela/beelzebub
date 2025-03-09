package plugins

import (
	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
	"github.com/mariocandela/beelzebub/v3/tracer"
	"github.com/stretchr/testify/assert"
	"net/http"
	"os"
	"testing"
)

const SystemPromptLen = 4

func TestBuildPromptEmptyHistory(t *testing.T) {
	//Given
	var histories []Message
	command := "pwd"

	honeypot := LLMHoneypot{
		Histories: histories,
		Protocol:  tracer.SSH,
	}

	//When
	prompt, err := honeypot.buildPrompt(command)

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

	honeypot := LLMHoneypot{
		Histories: histories,
		Protocol:  tracer.SSH,
	}

	//When
	prompt, err := honeypot.buildPrompt(command)

	//Then
	assert.Nil(t, err)
	assert.Equal(t, SystemPromptLen+1, len(prompt))
}

func TestBuildPromptWithCustomPrompt(t *testing.T) {
	//Given
	var histories = []Message{
		{
			Role:    "cat hello.txt",
			Content: "world",
		},
	}

	command := "pwd"

	honeypot := LLMHoneypot{
		Histories:    histories,
		Protocol:     tracer.SSH,
		CustomPrompt: "act as calculator",
	}

	//When
	prompt, err := honeypot.buildPrompt(command)

	//Then
	assert.Nil(t, err)
	assert.Equal(t, prompt[0].Content, "act as calculator")
	assert.Equal(t, prompt[0].Role, SYSTEM.String())
}

func TestBuildExecuteModelFailValidation(t *testing.T) {

	llmHoneypot := LLMHoneypot{
		Histories: make([]Message, 0),
		OpenAIKey: "",
		Protocol:  tracer.SSH,
		Model:     "gpt-4o",
		Provider:  OpenAI,
	}

	openAIGPTVirtualTerminal := InitLLMHoneypot(llmHoneypot)

	_, err := openAIGPTVirtualTerminal.ExecuteModel("test")

	assert.Equal(t, "openAIKey is empty", err.Error())
}

func TestBuildExecuteModelOpenAISecretKeyFromEnv(t *testing.T) {

	llmHoneypot := LLMHoneypot{
		Histories: make([]Message, 0),
		OpenAIKey: "",
		Protocol:  tracer.SSH,
		Model:     "gpt-4o",
		Provider:  OpenAI,
	}

	os.Setenv("OPEN_AI_SECRET_KEY", "sdjdnklfjndslkjanfk")

	openAIGPTVirtualTerminal := InitLLMHoneypot(llmHoneypot)

	assert.Equal(t, "sdjdnklfjndslkjanfk", openAIGPTVirtualTerminal.OpenAIKey)

}

func TestBuildExecuteModelWithCustomPrompt(t *testing.T) {
	client := resty.New()
	httpmock.ActivateNonDefault(client.GetClient())
	defer httpmock.DeactivateAndReset()

	// Given
	httpmock.RegisterMatcherResponder("POST", openAIEndpoint,
		httpmock.BodyContainsString("hello world"),
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
		Histories:    make([]Message, 0),
		OpenAIKey:    "sdjdnklfjndslkjanfk",
		Protocol:     tracer.HTTP,
		Model:        "gpt-4o",
		Provider:     OpenAI,
		CustomPrompt: "hello world",
	}

	openAIGPTVirtualTerminal := InitLLMHoneypot(llmHoneypot)
	openAIGPTVirtualTerminal.client = client

	//When
	str, err := openAIGPTVirtualTerminal.ExecuteModel("GET /.aws/credentials")

	//Then
	assert.Nil(t, err)
	assert.Equal(t, "[default]\nregion = us-west-2\noutput = json", str)
}

func TestBuildExecuteModelFailValidationStrategyType(t *testing.T) {

	llmHoneypot := LLMHoneypot{
		Histories: make([]Message, 0),
		OpenAIKey: "",
		Protocol:  tracer.TCP,
		Model:     "gpt-4o",
		Provider:  OpenAI,
	}

	openAIGPTVirtualTerminal := InitLLMHoneypot(llmHoneypot)

	_, err := openAIGPTVirtualTerminal.ExecuteModel("test")

	assert.Equal(t, "no prompt for protocol selected", err.Error())
}

func TestBuildExecuteModelFailValidationModelType(t *testing.T) {
	// Given
	llmHoneypot := LLMHoneypot{
		Histories: make([]Message, 0),
		Protocol:  tracer.SSH,
		Model:     "llama3",
		Provider:  5,
	}

	openAIGPTVirtualTerminal := InitLLMHoneypot(llmHoneypot)

	//When
	_, err := openAIGPTVirtualTerminal.ExecuteModel("ls")

	//Then
	assert.Errorf(t, err, "no model selected")
}

func TestBuildExecuteModelSSHWithResultsOpenAI(t *testing.T) {
	client := resty.New()
	httpmock.ActivateNonDefault(client.GetClient())
	defer httpmock.DeactivateAndReset()

	// Given
	httpmock.RegisterResponder("POST", openAIEndpoint,
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
		Model:     "gpt-4o",
		Provider:  OpenAI,
	}

	openAIGPTVirtualTerminal := InitLLMHoneypot(llmHoneypot)
	openAIGPTVirtualTerminal.client = client

	//When
	str, err := openAIGPTVirtualTerminal.ExecuteModel("ls")

	//Then
	assert.Nil(t, err)
	assert.Equal(t, "prova.txt", str)
}

func TestBuildExecuteModelSSHWithResultsLLama(t *testing.T) {
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
		Model:     "llama3",
		Provider:  Ollama,
	}

	openAIGPTVirtualTerminal := InitLLMHoneypot(llmHoneypot)
	openAIGPTVirtualTerminal.client = client

	//When
	str, err := openAIGPTVirtualTerminal.ExecuteModel("ls")

	//Then
	assert.Nil(t, err)
	assert.Equal(t, "prova.txt", str)
}

func TestBuildExecuteModelSSHWithoutResults(t *testing.T) {
	client := resty.New()
	httpmock.ActivateNonDefault(client.GetClient())
	defer httpmock.DeactivateAndReset()

	// Given
	httpmock.RegisterResponder("POST", openAIEndpoint,
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
		Model:     "gpt-4o",
		Provider:  OpenAI,
	}

	openAIGPTVirtualTerminal := InitLLMHoneypot(llmHoneypot)
	openAIGPTVirtualTerminal.client = client

	//When
	_, err := openAIGPTVirtualTerminal.ExecuteModel("ls")

	//Then
	assert.Equal(t, "no choices", err.Error())
}

func TestBuildExecuteModelHTTPWithResults(t *testing.T) {
	client := resty.New()
	httpmock.ActivateNonDefault(client.GetClient())
	defer httpmock.DeactivateAndReset()

	// Given
	httpmock.RegisterResponder("POST", openAIEndpoint,
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
		Model:     "gpt-4o",
		Provider:  OpenAI,
	}

	openAIGPTVirtualTerminal := InitLLMHoneypot(llmHoneypot)
	openAIGPTVirtualTerminal.client = client

	//When
	str, err := openAIGPTVirtualTerminal.ExecuteModel("GET /.aws/credentials")

	//Then
	assert.Nil(t, err)
	assert.Equal(t, "[default]\nregion = us-west-2\noutput = json", str)
}

func TestBuildExecuteModelHTTPWithoutResults(t *testing.T) {
	client := resty.New()
	httpmock.ActivateNonDefault(client.GetClient())
	defer httpmock.DeactivateAndReset()

	// Given
	httpmock.RegisterResponder("POST", openAIEndpoint,
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
		Model:     "gpt-4o",
		Provider:  OpenAI,
	}

	openAIGPTVirtualTerminal := InitLLMHoneypot(llmHoneypot)
	openAIGPTVirtualTerminal.client = client

	//When
	_, err := openAIGPTVirtualTerminal.ExecuteModel("GET /.aws/credentials")

	//Then
	assert.Equal(t, "no choices", err.Error())
}

func TestFromString(t *testing.T) {
	model, err := FromStringToLLMProvider("openai")
	assert.Nil(t, err)
	assert.Equal(t, OpenAI, model)

	model, err = FromStringToLLMProvider("ollama")
	assert.Nil(t, err)
	assert.Equal(t, Ollama, model)

	model, err = FromStringToLLMProvider("beelzebub-model")
	assert.Errorf(t, err, "provider beelzebub-model not found")
}

func TestBuildExecuteModelSSHWithoutPlaintextSection(t *testing.T) {
	client := resty.New()
	httpmock.ActivateNonDefault(client.GetClient())
	defer httpmock.DeactivateAndReset()

	// Given
	httpmock.RegisterResponder("POST", ollamaEndpoint,
		func(req *http.Request) (*http.Response, error) {
			resp, err := httpmock.NewJsonResponse(200, &Response{
				Message: Message{
					Role:    SYSTEM.String(),
					Content: "```plaintext\n```\n",
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
		Model:     "llama3",
	}

	openAIGPTVirtualTerminal := InitLLMHoneypot(llmHoneypot)
	openAIGPTVirtualTerminal.client = client

	//When
	str, err := openAIGPTVirtualTerminal.ExecuteModel("ls")

	//Then
	assert.Nil(t, err)
	assert.Equal(t, "", str)
}

func TestBuildExecuteModelSSHWithoutQuotesSection(t *testing.T) {
	client := resty.New()
	httpmock.ActivateNonDefault(client.GetClient())
	defer httpmock.DeactivateAndReset()

	// Given
	httpmock.RegisterResponder("POST", ollamaEndpoint,
		func(req *http.Request) (*http.Response, error) {
			resp, err := httpmock.NewJsonResponse(200, &Response{
				Message: Message{
					Role:    SYSTEM.String(),
					Content: "```\n```\n",
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
		Model:     "llama3",
		Provider:  Ollama,
	}

	openAIGPTVirtualTerminal := InitLLMHoneypot(llmHoneypot)
	openAIGPTVirtualTerminal.client = client

	//When
	str, err := openAIGPTVirtualTerminal.ExecuteModel("ls")

	//Then
	assert.Nil(t, err)
	assert.Equal(t, "", str)
}

func TestRemoveQuotes(t *testing.T) {
	plaintext := "```plaintext\n```"
	bash := "```bash\n```"
	onlyQuotes := "```\n```"
	complexText := "```plaintext\ntop - 10:30:48 up 1 day,  4:30,  2 users,  load average: 0.15, 0.10, 0.08\nTasks: 198 total,   1 running, 197 sleeping,   0 stopped,   0 zombie\n```"
	complexText2 := "```\ntop - 15:06:59 up 10 days,  3:17,  1 user,  load average: 0.10, 0.09, 0.08\nTasks: 285 total\n```"

	assert.Equal(t, "", removeQuotes(plaintext))
	assert.Equal(t, "", removeQuotes(bash))
	assert.Equal(t, "", removeQuotes(onlyQuotes))
	assert.Equal(t, "top - 10:30:48 up 1 day,  4:30,  2 users,  load average: 0.15, 0.10, 0.08\nTasks: 198 total,   1 running, 197 sleeping,   0 stopped,   0 zombie\n", removeQuotes(complexText))
	assert.Equal(t, "top - 15:06:59 up 10 days,  3:17,  1 user,  load average: 0.10, 0.09, 0.08\nTasks: 285 total\n", removeQuotes(complexText2))
}
