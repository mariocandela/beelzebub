package TCP

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/mariocandela/beelzebub/v3/historystore"
	"github.com/mariocandela/beelzebub/v3/parser"
	"github.com/mariocandela/beelzebub/v3/plugins"
	"github.com/mariocandela/beelzebub/v3/tracer"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type TCPStrategy struct {
	Sessions *historystore.HistoryStore
}

func (tcpStrategy *TCPStrategy) Init(servConf parser.BeelzebubServiceConfiguration, tr tracer.Tracer) error {
	if tcpStrategy.Sessions == nil {
		tcpStrategy.Sessions = historystore.NewHistoryStore()
	}
	go tcpStrategy.Sessions.HistoryCleaner()

	listen, err := net.Listen("tcp", servConf.Address)
	if err != nil {
		log.Errorf("Error during init TCP Protocol: %s", err.Error())
		return err
	}

	go func() {
		for {
			if conn, err := listen.Accept(); err == nil {
				go func(c net.Conn) {
					defer func() {
						if r := recover(); r != nil {
							log.Errorf("panic in TCP handler: %v", r)
						}
					}()
					handleTCPConnection(c, servConf, tr, tcpStrategy)
				}(conn)
			}
		}
	}()

	log.WithFields(log.Fields{
		"port":     servConf.Address,
		"banner":   servConf.Banner,
		"commands": len(servConf.Commands),
	}).Infof("Init service %s", servConf.Protocol)
	return nil
}

func handleTCPConnection(conn net.Conn, servConf parser.BeelzebubServiceConfiguration, tr tracer.Tracer, tcpStrategy *TCPStrategy) {
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(time.Duration(servConf.DeadlineTimeoutSeconds) * time.Second))

	host, port, _ := net.SplitHostPort(conn.RemoteAddr().String())

	// Send banner if configured
	if servConf.Banner != "" {
		conn.Write(fmt.Appendf([]byte{}, "%s\n", servConf.Banner))
	}

	// Backward compatibility: if no commands configured, use legacy behavior
	if len(servConf.Commands) == 0 {
		buffer := make([]byte, 1024)
		command := ""

		if n, err := conn.Read(buffer); err == nil {
			command = string(buffer[:n])
		}

		tr.TraceEvent(tracer.Event{
			Msg:         "New TCP attempt",
			Protocol:    tracer.TCP.String(),
			Command:     command,
			Status:      tracer.Stateless.String(),
			RemoteAddr:  conn.RemoteAddr().String(),
			SourceIp:    host,
			SourcePort:  port,
			ID:          uuid.New().String(),
			Description: servConf.Description,
		})
		return
	}

	// Interactive session mode
	sessionID := uuid.New()
	sessionKey := "TCP" + host

	tr.TraceEvent(tracer.Event{
		Msg:         "New TCP Session",
		Protocol:    tracer.TCP.String(),
		RemoteAddr:  conn.RemoteAddr().String(),
		SourceIp:    host,
		SourcePort:  port,
		Status:      tracer.Start.String(),
		ID:          sessionID.String(),
		Description: servConf.Description,
	})

	// Load history for LLM context
	var histories []plugins.Message
	if tcpStrategy.Sessions.HasKey(sessionKey) {
		histories = tcpStrategy.Sessions.Query(sessionKey)
	}

	// Interactive command loop
	for {
		buffer := make([]byte, 4096)
		n, err := conn.Read(buffer)
		if err != nil {
			break
		}

		commandInput := strings.TrimRight(string(buffer[:n]), "\r\n")

		// Match command against regexes
		matched := false
		for _, command := range servConf.Commands {
			if command.Regex.MatchString(commandInput) {
				matched = true
				commandOutput := command.Handler
				handlerName := command.Name
				if handlerName == "" {
					handlerName = "configured_regex"
				}

				// LLM integration
				if command.Plugin == plugins.LLMPluginName {
					llmProvider, err := plugins.FromStringToLLMProvider(servConf.Plugin.LLMProvider)
					if err != nil {
						log.Errorf("error: %s, fallback OpenAI", err.Error())
						llmProvider = plugins.OpenAI
					}
					llmHoneypot := plugins.BuildHoneypot(histories, tracer.TCP, llmProvider, servConf)
					llmHoneypotInstance := plugins.InitLLMHoneypot(*llmHoneypot)
					if commandOutput, err = llmHoneypotInstance.ExecuteModel(commandInput, host); err != nil {
						log.Errorf("error ExecuteModel: %s, %s", commandInput, err.Error())
						commandOutput = ""
					}
				}

				// Store command and response in history
				var newEntries []plugins.Message
				newEntries = append(newEntries, plugins.Message{Role: plugins.USER.String(), Content: commandInput})
				newEntries = append(newEntries, plugins.Message{Role: plugins.ASSISTANT.String(), Content: commandOutput})
				tcpStrategy.Sessions.Append(sessionKey, newEntries...)
				histories = append(histories, newEntries...)

				// Send response to client
				if commandOutput != "" {
					_, err := conn.Write([]byte(commandOutput))
					if err != nil {
						break
					}
				}

				// Trace interaction event
				tr.TraceEvent(tracer.Event{
					Msg:           "TCP Session Interaction",
					RemoteAddr:    conn.RemoteAddr().String(),
					SourceIp:      host,
					SourcePort:    port,
					Status:        tracer.Interaction.String(),
					Command:       commandInput,
					CommandOutput: commandOutput,
					ID:            sessionID.String(),
					Protocol:      tracer.TCP.String(),
					Description:   servConf.Description,
					Handler:       handlerName,
				})

				break
			}
		}

		// If no command matched
		if !matched {
			tr.TraceEvent(tracer.Event{
				Msg:         "TCP Session Interaction",
				RemoteAddr:  conn.RemoteAddr().String(),
				SourceIp:    host,
				SourcePort:  port,
				Status:      tracer.Interaction.String(),
				Command:     commandInput,
				ID:          sessionID.String(),
				Protocol:    tracer.TCP.String(),
				Description: servConf.Description,
				Handler:     "not_found",
			})
		}
	}

	// Trace session end
	tr.TraceEvent(tracer.Event{
		Msg:      "End TCP Session",
		Status:   tracer.End.String(),
		ID:       sessionID.String(),
		Protocol: tracer.TCP.String(),
	})
}
