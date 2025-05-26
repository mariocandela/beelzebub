package plugins

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
	"github.com/mariocandela/beelzebub/v3/parser"
	"github.com/mariocandela/beelzebub/v3/tracer"
	"github.com/stretchr/testify/assert"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestBuildSendEventFailValidation(t *testing.T) {
	beelzebubCloud := InitBeelzebubCloud("", "", false)

	_, err := beelzebubCloud.SendEvent(tracer.Event{})

	assert.Equal(t, "authToken is empty", err.Error())
}

func TestBuildSendEventWithResults(t *testing.T) {
	client := resty.New()
	httpmock.ActivateNonDefault(client.GetClient())
	defer httpmock.DeactivateAndReset()

	uri := "localhost:8081"

	// Given
	httpmock.RegisterResponder("POST", fmt.Sprintf("%s/events", uri),
		func(req *http.Request) (*http.Response, error) {
			resp, err := httpmock.NewJsonResponse(200, &tracer.Event{})
			if err != nil {
				return httpmock.NewStringResponse(500, ""), nil
			}
			return resp, nil
		},
	)

	beelzebubCloud := InitBeelzebubCloud(uri, "sdjdnklfjndslkjanfk", false)
	beelzebubCloud.client = client

	//When
	result, err := beelzebubCloud.SendEvent(tracer.Event{})

	//Then
	assert.Equal(t, true, result)
	assert.Nil(t, err)
}

func TestBuildSendEventErro(t *testing.T) {
	client := resty.New()
	httpmock.ActivateNonDefault(client.GetClient())
	defer httpmock.DeactivateAndReset()

	uri := "localhost:8081/events"

	// Given
	httpmock.RegisterResponder("POST", uri,
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewStringResponse(500, ""), nil
		},
	)

	beelzebubCloud := InitBeelzebubCloud(uri, "sdjdnklfjndslkjanfk", false)
	beelzebubCloud.client = client

	//When
	result, _ := beelzebubCloud.SendEvent(tracer.Event{})

	//Then
	assert.Equal(t, false, result)
}

func TestGetHoneypotsConfigurationsWithResults(t *testing.T) {
	client := resty.New()
	httpmock.ActivateNonDefault(client.GetClient())
	defer httpmock.DeactivateAndReset()

	uri := "localhost:8081"

	// Given
	httpmock.RegisterResponder("GET", fmt.Sprintf("%s/honeypots", uri),
		func(req *http.Request) (*http.Response, error) {
			resp, err := httpmock.NewJsonResponse(200, &[]HoneypotConfigResponseDTO{
				{
					ID:      "123456",
					Config:  "apiVersion: \"v1\"\nprotocol: \"ssh\"\naddress: \":2222\"\ndescription: \"SSH interactive ChatGPT\"\ncommands:\n  - regex: \"^(.+)$\"\n    plugin: \"LLMHoneypot\"\nserverVersion: \"OpenSSH\"\nserverName: \"ubuntu\"\npasswordRegex: \"^(root|qwerty|Smoker666|123456|jenkins|minecraft|sinus|alex|postgres|Ly123456)$\"\ndeadlineTimeoutSeconds: 60\nplugin:\n  llmModel: \"gpt-4o\"\n  openAISecretKey: \"1234\"\n",
					TokenID: "1234567",
				},
			})
			if err != nil {
				return httpmock.NewStringResponse(500, ""), nil
			}
			return resp, nil
		},
	)

	beelzebubCloud := InitBeelzebubCloud(uri, "sdjdnklfjndslkjanfk", false)
	beelzebubCloud.client = client

	//When
	result, err := beelzebubCloud.GetHoneypotsConfigurations()

	//Then
	assert.Equal(t, &[]parser.BeelzebubServiceConfiguration{
		{
			ApiVersion:  "v1",
			Protocol:    "ssh",
			Address:     ":2222",
			Description: "SSH interactive ChatGPT",
			Commands: []parser.Command{
				{
					Regex:  "^(.+)$",
					Plugin: "LLMHoneypot",
				},
			},
			ServerVersion:          "OpenSSH",
			ServerName:             "ubuntu",
			PasswordRegex:          "^(root|qwerty|Smoker666|123456|jenkins|minecraft|sinus|alex|postgres|Ly123456)$",
			DeadlineTimeoutSeconds: 60,
			Plugin: parser.Plugin{
				LLMModel:        "gpt-4o",
				OpenAISecretKey: "1234",
			},
		},
	}, &result)
	assert.Equal(t, "5a76814e57a6c6ab48da4380f6fec988efc8dc6e51a64d78491d974430b773ce", beelzebubCloud.ConfigurationsHash.String())
	assert.Nil(t, err)
}

func TestGetHoneypotsConfigurationsWithErrorValidation(t *testing.T) {
	//Given
	beelzebubCloud := InitBeelzebubCloud("", "", false)

	//When
	result, err := beelzebubCloud.GetHoneypotsConfigurations()

	//Then
	assert.Nil(t, result)
	assert.Equal(t, "authToken is empty", err.Error())
}

func TestGetHoneypotsConfigurationsWithErrorAPI(t *testing.T) {
	client := resty.New()
	httpmock.ActivateNonDefault(client.GetClient())
	defer httpmock.DeactivateAndReset()

	uri := "localhost:8081"

	// Given
	httpmock.RegisterResponder("GET", fmt.Sprintf("%s/honeypots", uri),
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewStringResponse(500, ""), nil
		},
	)

	beelzebubCloud := InitBeelzebubCloud(uri, "sdjdnklfjndslkjanfk", false)
	beelzebubCloud.client = client

	//When
	result, err := beelzebubCloud.GetHoneypotsConfigurations()

	//Then
	assert.Nil(t, result)
	assert.Equal(t, "Response code: 500, error: ", err.Error())
}

func TestGetHoneypotsConfigurationsWithErrorUnmarshal(t *testing.T) {
	client := resty.New()
	httpmock.ActivateNonDefault(client.GetClient())
	defer httpmock.DeactivateAndReset()

	uri := "localhost:8081"

	// Given
	httpmock.RegisterResponder("GET", fmt.Sprintf("%s/honeypots", uri),
		func(req *http.Request) (*http.Response, error) {
			resp, err := httpmock.NewJsonResponse(200, "error")
			if err != nil {
				return httpmock.NewStringResponse(500, ""), nil
			}
			return resp, nil
		},
	)

	beelzebubCloud := InitBeelzebubCloud(uri, "sdjdnklfjndslkjanfk", false)
	beelzebubCloud.client = client

	//When
	result, err := beelzebubCloud.GetHoneypotsConfigurations()

	//Then
	assert.Nil(t, result)
	assert.Equal(t, "json: cannot unmarshal string into Go value of type []plugins.HoneypotConfigResponseDTO", err.Error())
}

func TestGetHoneypotsConfigurationsWithErrorDeserializeYaml(t *testing.T) {
	client := resty.New()
	httpmock.ActivateNonDefault(client.GetClient())
	defer httpmock.DeactivateAndReset()

	uri := "localhost:8081"

	// Given
	httpmock.RegisterResponder("GET", fmt.Sprintf("%s/honeypots", uri),
		func(req *http.Request) (*http.Response, error) {
			resp, err := httpmock.NewJsonResponse(200, &[]HoneypotConfigResponseDTO{
				{
					ID:      "123456",
					Config:  "error",
					TokenID: "1234567",
				},
			})
			if err != nil {
				return httpmock.NewStringResponse(500, ""), nil
			}
			return resp, nil
		},
	)

	beelzebubCloud := InitBeelzebubCloud(uri, "sdjdnklfjndslkjanfk", false)
	beelzebubCloud.client = client

	//When
	result, err := beelzebubCloud.GetHoneypotsConfigurations()

	//Then
	assert.Nil(t, result)
	assert.Equal(t, "yaml: unmarshal errors:\n  line 1: cannot unmarshal !!str `error` into parser.BeelzebubServiceConfiguration", err.Error())
}

func TestVerifyConfigurationsChanged(t *testing.T) {
	client := resty.New()
	httpmock.ActivateNonDefault(client.GetClient())
	defer httpmock.DeactivateAndReset()

	uri := "localhost:8081"

	// Given
	httpmock.RegisterResponder("GET", fmt.Sprintf("%s/honeypots", uri),
		func(req *http.Request) (*http.Response, error) {
			resp, err := httpmock.NewJsonResponse(200, &[]HoneypotConfigResponseDTO{
				{
					ID:      "123456",
					Config:  "apiVersion: \"v1\"\nprotocol: \"ssh\"\naddress: \":2222\"\ndescription: \"SSH interactive ChatGPT\"\ncommands:\n  - regex: \"^(.+)$\"\n    plugin: \"LLMHoneypot\"\nserverVersion: \"OpenSSH\"\nserverName: \"ubuntu\"\npasswordRegex: \"^(root|qwerty|Smoker666|123456|jenkins|minecraft|sinus|alex|postgres|Ly123456)$\"\ndeadlineTimeoutSeconds: 60\nplugin:\n  llmModel: \"gpt4-o\"\n  openAISecretKey: \"1234\"\n",
					TokenID: "1234567",
				},
			})
			if err != nil {
				return httpmock.NewStringResponse(500, ""), nil
			}
			return resp, nil
		},
	)

	hashConfigurations := strings.Builder{}
	hashConfigurations.WriteString("sdafsdgfhggfdfdgdsgdfgsdfg")
	var exitInvoked bool = false

	exitCalled := make(chan bool)

	// Override exitFunction to set the flag AND exit the infinite loop with panic/recover
	exitFunction = func(c int) {
		exitInvoked = true
		exitCalled <- true
		panic("Exit function called")
	}

	beelzebubCloud := InitBeelzebubCloud(uri, "sdjdnklfjndslkjanfk", false)
	beelzebubCloud.client = client
	beelzebubCloud.ConfigurationsHash = hashConfigurations

	// Run verifyConfigurationsChanged in a goroutine with panic recovery
	go func() {
		defer func() {
			// Recover from the panic
			recover()
		}()
		err := beelzebubCloud.verifyConfigurationsChanged()
		t.Errorf("verifyConfigurationsChanged returned unexpectedly with error: %v", err)
	}()

	// Wait for exitFunction to be called or timeout
	select {
	case <-exitCalled:
		// exitFunction was called, test should pass
	case <-time.After(5 * time.Second):
		t.Fatal("Test timed out waiting for exitFunction to be called")
	}

	// Then
	assert.Equal(t, true, exitInvoked)
}

func TestMapToEventDTO(t *testing.T) {
	event := tracer.Event{
		DateTime:        "2025-05-01T16:18:13Z",
		RemoteAddr:      "1.1.1.1:12345",
		Protocol:        "SSH",
		Command:         "cd /tmp",
		CommandOutput:   "",
		Status:          "Interaction",
		Msg:             "New SSH Terminal Session",
		ID:              "4f104892-738f-47ac-950f-6afce1b742c7",
		Environ:         "qwerty",
		User:            "root",
		Password:        "root",
		Client:          "ssh",
		Headers:         map[string][]string{"Host": {"beelzebub-honeypot.com"}},
		Cookies:         "qwerty",
		UserAgent:       "qwerty",
		HostHTTPRequest: "beelzebub-honeypot.com",
		Body:            "qwerty",
		HTTPMethod:      "GET",
		RequestURI:      "/qwerty",
		Description:     "qwerty",
		SourceIp:        "1.1.1.1",
		SourcePort:      "12345",
		TLSServerName:   "beelzebub-honeypot.com",
	}
	beelzebubCloud := InitBeelzebubCloud("localhost:8081", "sdjdnklfjndslkjanfk", false)
	eventDTO, err := beelzebubCloud.mapToEventDTO(event)
	assert.Nil(t, err)

	assert.Equal(t, EventDTO{
		DateTime:        "2025-05-01T16:18:13Z",
		RemoteAddr:      "1.1.1.1:12345",
		Protocol:        "SSH",
		Command:         "cd /tmp",
		CommandOutput:   "",
		Status:          "Interaction",
		Msg:             "New SSH Terminal Session",
		ID:              "4f104892-738f-47ac-950f-6afce1b742c7",
		Environ:         "qwerty",
		User:            "root",
		Password:        "root",
		Client:          "ssh",
		Headers:         "{\"Host\":[\"beelzebub-honeypot.com\"]}",
		Cookies:         "qwerty",
		UserAgent:       "qwerty",
		HostHTTPRequest: "beelzebub-honeypot.com",
		Body:            "qwerty",
		HTTPMethod:      "GET",
		RequestURI:      "/qwerty",
		Description:     "qwerty",
		SourceIp:        "1.1.1.1",
		SourcePort:      "12345",
		TLSServerName:   "beelzebub-honeypot.com",
	}, eventDTO)
}
