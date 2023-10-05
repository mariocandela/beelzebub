package strategies

import (
	"fmt"
	"github.com/mariocandela/beelzebub/parser"
	"github.com/mariocandela/beelzebub/tracer"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type HTTPStrategy struct {
	beelzebubServiceConfiguration parser.BeelzebubServiceConfiguration
}

func (httpStrategy HTTPStrategy) Init(beelzebubServiceConfiguration parser.BeelzebubServiceConfiguration, tr tracer.Tracer) error {
	httpStrategy.beelzebubServiceConfiguration = beelzebubServiceConfiguration
	serverMux := http.NewServeMux()

	serverMux.HandleFunc("/", func(responseWriter http.ResponseWriter, request *http.Request) {
		traceRequest(request, tr, beelzebubServiceConfiguration.Description)
		for _, command := range httpStrategy.beelzebubServiceConfiguration.Commands {
			matched, err := regexp.MatchString(command.Regex, request.RequestURI)
			if err != nil {
				log.Errorf("Error regex: %s, %s", command.Regex, err.Error())
				continue
			}

			if matched {
				setResponseHeaders(responseWriter, command.Headers, command.StatusCode)
				fmt.Fprintf(responseWriter, command.Handler)
				break
			}
		}
	})
	go func() {
		err := http.ListenAndServe(httpStrategy.beelzebubServiceConfiguration.Address, serverMux)
		if err != nil {
			log.Errorf("Error during init HTTP Protocol: %s", err.Error())
			return
		}
	}()

	log.WithFields(log.Fields{
		"port":     beelzebubServiceConfiguration.Address,
		"commands": len(beelzebubServiceConfiguration.Commands),
	}).Infof("Init service: %s", beelzebubServiceConfiguration.Description)
	return nil
}

func traceRequest(request *http.Request, tr tracer.Tracer, HoneypotDescription string) {
	bodyBytes, err := io.ReadAll(request.Body)
	body := ""
	if err == nil {
		body = string(bodyBytes)
	}
	tr.TraceEvent(tracer.Event{
		Msg:             "HTTP New request",
		RequestURI:      request.RequestURI,
		Protocol:        tracer.HTTP.String(),
		HTTPMethod:      request.Method,
		Body:            body,
		HostHTTPRequest: request.Host,
		UserAgent:       request.UserAgent(),
		Cookies:         mapCookiesToString(request.Cookies()),
		Headers:         mapHeaderToString(request.Header),
		Status:          tracer.Stateless.String(),
		RemoteAddr:      request.RemoteAddr,
		ID:              uuid.New().String(),
		Description:     HoneypotDescription,
	})
}

func mapHeaderToString(headers http.Header) string {
	headersString := ""

	for key := range headers {
		for _, values := range headers[key] {
			headersString += fmt.Sprintf("[Key: %s, values: %s],", key, values)
		}
	}

	return headersString
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
