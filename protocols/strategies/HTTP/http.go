package HTTP

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"regexp"
	"strings"

	"github.com/mariocandela/beelzebub/v3/parser"
	"github.com/mariocandela/beelzebub/v3/plugins"
	"github.com/mariocandela/beelzebub/v3/tracer"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type HTTPStrategy struct {
	beelzebubServiceConfiguration parser.BeelzebubServiceConfiguration
}

type httpResponse struct {
	StatusCode int
	Headers    []string
	Body       string
}

func (httpStrategy HTTPStrategy) Init(beelzebubServiceConfiguration parser.BeelzebubServiceConfiguration, tr tracer.Tracer) error {
	httpStrategy.beelzebubServiceConfiguration = beelzebubServiceConfiguration
	serverMux := http.NewServeMux()

	serverMux.HandleFunc("/", func(responseWriter http.ResponseWriter, request *http.Request) {
		traceRequest(request, tr, beelzebubServiceConfiguration.Description)
		var matched bool
		var resp httpResponse
		var err error
		for _, command := range httpStrategy.beelzebubServiceConfiguration.Commands {
			var err error
			matched, err = regexp.MatchString(command.Regex, request.RequestURI)
			if err != nil {
				log.Errorf("Error regex: %s, %s", command.Regex, err.Error())
				continue
			}

			if matched {
				resp, err = buildHTTPResponse(beelzebubServiceConfiguration, command, request)
				if err != nil {
					log.Errorf("error building http response: %s: %v", request.RequestURI, err)
				}
				break
			}
		}
		// If none of the main commands matched, and we have a fallback command configured, process it here.
		// The regexp is ignored for fallback commands, as they are catch-all for any request.
		if !matched && httpStrategy.beelzebubServiceConfiguration.FallbackCommand.Handler != "" {
			command := httpStrategy.beelzebubServiceConfiguration.FallbackCommand
			resp, err = buildHTTPResponse(beelzebubServiceConfiguration, command, request)
			if err != nil {
				log.Errorf("error building http response: %s: %v", request.RequestURI, err)
			}
		}
		setResponseHeaders(responseWriter, resp.Headers, resp.StatusCode)
		fmt.Fprint(responseWriter, resp.Body)

	})
	go func() {
		var err error
		// Launch a TLS supporting server if we are supplied a TLS Key and Certificate.
		// If relative paths are supplied, they are relative to the CWD of the binary.
		// The can be self-signed, only the client will validate this (or not).
		if httpStrategy.beelzebubServiceConfiguration.TLSKeyPath != "" && httpStrategy.beelzebubServiceConfiguration.TLSCertPath != "" {
			err = http.ListenAndServeTLS(
				httpStrategy.beelzebubServiceConfiguration.Address,
				httpStrategy.beelzebubServiceConfiguration.TLSCertPath,
				httpStrategy.beelzebubServiceConfiguration.TLSKeyPath,
				serverMux)
		} else {
			err = http.ListenAndServe(httpStrategy.beelzebubServiceConfiguration.Address, serverMux)
		}
		if err != nil {
			log.Errorf("error during init HTTP Protocol: %s", err.Error())
			return
		}
	}()

	log.WithFields(log.Fields{
		"port":     beelzebubServiceConfiguration.Address,
		"commands": len(beelzebubServiceConfiguration.Commands),
	}).Infof("Init service: %s", beelzebubServiceConfiguration.Description)
	return nil
}

func buildHTTPResponse(beelzebubServiceConfiguration parser.BeelzebubServiceConfiguration, command parser.Command, request *http.Request) (httpResponse, error) {
	resp := httpResponse{
		Body:       command.Handler,
		Headers:    command.Headers,
		StatusCode: command.StatusCode,
	}

	if command.Plugin == plugins.LLMPluginName {
		llmProvider, err := plugins.FromStringToLLMProvider(beelzebubServiceConfiguration.Plugin.LLMProvider)
		if err != nil {
			log.Errorf("Error: %s", err.Error())
			resp.Body = "404 Not Found!"
		}

		llmHoneypot := plugins.LLMHoneypot{
			Histories:    make([]plugins.Message, 0),
			OpenAIKey:    beelzebubServiceConfiguration.Plugin.OpenAISecretKey,
			Protocol:     tracer.HTTP,
			Host:         beelzebubServiceConfiguration.Plugin.Host,
			Model:        beelzebubServiceConfiguration.Plugin.LLMModel,
			Provider:     llmProvider,
			CustomPrompt: beelzebubServiceConfiguration.Plugin.Prompt,
		}
		llmHoneypotInstance := plugins.InitLLMHoneypot(llmHoneypot)

		command := fmt.Sprintf("%s %s", request.Method, request.RequestURI)

		if completions, err := llmHoneypotInstance.ExecuteModel(command); err != nil {
			log.Errorf("Error ExecuteModel: %s, %s", command, err.Error())
			resp.Body = "404 Not Found!"
		} else {
			resp.Body = completions
		}
	}
	return resp, nil
}

func traceRequest(request *http.Request, tr tracer.Tracer, HoneypotDescription string) {
	bodyBytes, err := io.ReadAll(request.Body)
	body := ""
	if err == nil {
		body = string(bodyBytes)
	}
	host, port, _ := net.SplitHostPort(request.RemoteAddr)

	event := tracer.Event{
		Msg:             "HTTP New request",
		RequestURI:      request.RequestURI,
		Protocol:        tracer.HTTP.String(),
		HTTPMethod:      request.Method,
		Body:            body,
		HostHTTPRequest: request.Host,
		UserAgent:       request.UserAgent(),
		Cookies:         mapCookiesToString(request.Cookies()),
		Headers:         request.Header,
		Status:          tracer.Stateless.String(),
		RemoteAddr:      request.RemoteAddr,
		SourceIp:        host,
		SourcePort:      port,
		ID:              uuid.New().String(),
		Description:     HoneypotDescription,
	}
	// Capture the TLS details from the request, if provided.
	if request.TLS != nil {
		event.Msg = "HTTPS New Request"
		event.TLSServerName = request.TLS.ServerName
	}
	tr.TraceEvent(event)
}

func mapCookiesToString(cookies []*http.Cookie) string {
	cookiesString := ""

	for _, cookie := range cookies {
		cookiesString += cookie.String()
	}

	return cookiesString
}

func setResponseHeaders(responseWriter http.ResponseWriter, headers []string, statusCode int) {
	for _, headerStr := range headers {
		keyValue := strings.Split(headerStr, ":")
		if len(keyValue) > 1 {
			responseWriter.Header().Add(keyValue[0], keyValue[1])
		}
	}
	// http.StatusText(statusCode): empty string if the code is unknown.
	if len(http.StatusText(statusCode)) > 0 {
		responseWriter.WriteHeader(statusCode)
	}
}
