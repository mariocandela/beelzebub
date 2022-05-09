package protocols

import (
	"beelzebub/parser"
	"beelzebub/tracer"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"regexp"
	"strings"
)

type HypertextTransferProtocolStrategy struct {
	serverMux                     *http.ServeMux
	beelzebubServiceConfiguration parser.BeelzebubServiceConfiguration
}

func (httpStrategy HypertextTransferProtocolStrategy) Init(beelzebubServiceConfiguration parser.BeelzebubServiceConfiguration, tr tracer.Tracer) error {
	httpStrategy.beelzebubServiceConfiguration = beelzebubServiceConfiguration
	httpStrategy.serverMux = http.NewServeMux()

	httpStrategy.buildHandler()

	go func() {
		httpStrategy.listenAndServe()
	}()

	log.WithFields(log.Fields{
		"port":     beelzebubServiceConfiguration.Address,
		"commands": len(beelzebubServiceConfiguration.Commands),
	}).Infof("Init service %s", beelzebubServiceConfiguration.Protocol)
	return nil
}

func (httpStrategy HypertextTransferProtocolStrategy) listenAndServe() {
	err := http.ListenAndServe(httpStrategy.beelzebubServiceConfiguration.Address, httpStrategy.serverMux)
	if err != nil {
		log.Errorf("Error during init HTTP Protocol: %s", err.Error())
		return
	}
}

func (httpStrategy HypertextTransferProtocolStrategy) buildHandler() {
	httpStrategy.serverMux.HandleFunc("/", func(responseWriter http.ResponseWriter, request *http.Request) {
		traceRequest(request)
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
}

func traceRequest(request *http.Request) {
	bodyBytes, err := io.ReadAll(request.Body)
	body := ""
	if err == nil {
		body = string(bodyBytes)
	}
	log.WithFields(log.Fields{
		"requestURI": request.RequestURI,
		"proto":      request.Proto,
		"method":     request.Method,
		"body":       body,
		"host":       request.Host,
		"userAgent":  request.UserAgent(),
		"cookies":    request.Cookies(),
		"ip":         request.RemoteAddr,
		"headers":    request.Header,
		"remoteAddr": request.RemoteAddr,
	}).Info("New HTTP request")
}

func setResponseHeaders(responseWriter http.ResponseWriter, headers []string, statusCode int) {
	responseWriter.WriteHeader(statusCode)
	for _, headerStr := range headers {
		keyValue := strings.Split(headerStr, ":")
		if len(keyValue) > 1 {
			responseWriter.Header().Add(keyValue[0], keyValue[1])
		}
	}
}
