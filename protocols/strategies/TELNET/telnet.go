package TELNET

import (
	"bytes"
	"fmt"
	"net"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"

	"github.com/mariocandela/beelzebub/v3/historystore"
	"github.com/mariocandela/beelzebub/v3/parser"
	"github.com/mariocandela/beelzebub/v3/plugins"
	"github.com/mariocandela/beelzebub/v3/tracer"
)

// Telnet IAC (Interpret As Command) constants
const (
	IAC               = 255 // Interpret As Command
	WILL              = 251 // Will perform option
	WONT              = 252 // Won't perform option
	ECHO              = 1   // Echo option
	SUPPRESS_GO_AHEAD = 3   // Suppress Go Ahead option
)

type TelnetStrategy struct {
	Sessions *historystore.HistoryStore
}

func (telnetStrategy *TelnetStrategy) Init(servConf parser.BeelzebubServiceConfiguration, tr tracer.Tracer) error {
	if telnetStrategy.Sessions == nil {
		telnetStrategy.Sessions = historystore.NewHistoryStore()
	}
	go telnetStrategy.Sessions.HistoryCleaner()

	go func() {
		listener, err := net.Listen("tcp4", servConf.Address)
		if err != nil {
			log.Errorf("error during init TELNET Protocol: %s", err.Error())
			return
		}
		defer listener.Close()

		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Errorf("error accepting TELNET connection: %s", err.Error())
				continue
			}

			// Set deadline timeout
			conn.SetDeadline(time.Now().Add(time.Duration(servConf.DeadlineTimeoutSeconds) * time.Second))

			go func(c net.Conn) {
				defer func() {
					if r := recover(); r != nil {
						log.Errorf("panic in TELNET handler: %v", r)
					}
				}()
				handleTelnetConnection(c, servConf, tr, telnetStrategy)
			}(conn)
		}
	}()

	log.WithFields(log.Fields{
		"port":     servConf.Address,
		"commands": len(servConf.Commands),
	}).Infof("GetInstance service %s", servConf.Protocol)
	return nil
}

func handleTelnetConnection(conn net.Conn, servConf parser.BeelzebubServiceConfiguration, tr tracer.Tracer, telnetStrategy *TelnetStrategy) {
	defer conn.Close()

	host, port, _ := net.SplitHostPort(conn.RemoteAddr().String())

	// Negotiate telnet options
	negotiateTelnet(conn)

	// Read and discard any client responses to our IAC
	drainInitialInput(conn)

	// Authentication phase
	_, err := conn.Write([]byte("login: \r\n"))
	if err != nil {
		return
	}

	username, err := readLine(conn)
	if err != nil {
		return
	}
	username = strings.TrimSpace(username)

	// Send password prompt with echo suppression
	_, err = conn.Write([]byte{IAC, WILL, ECHO})
	if err != nil {
		return
	}
	_, err = conn.Write([]byte("Password: "))
	if err != nil {
		return
	}

	password, err := readLine(conn)
	if err != nil {
		return
	}
	password = strings.TrimSpace(password)

	// Re-enable echo
	_, err = conn.Write([]byte{IAC, WONT, ECHO, '\r', '\n'})
	if err != nil {
		return
	}

	// Trace authentication attempt
	tr.TraceEvent(tracer.Event{
		Msg:         "New TELNET Login Attempt",
		Protocol:    tracer.TELNET.String(),
		Status:      tracer.Stateless.String(),
		User:        username,
		Password:    password,
		RemoteAddr:  conn.RemoteAddr().String(),
		SourceIp:    host,
		SourcePort:  port,
		ID:          uuid.New().String(),
		Description: servConf.Description,
	})

	// Validate password
	matched, err := regexp.MatchString(servConf.PasswordRegex, password)
	if err != nil {
		log.Errorf("error regex: %s, %s", servConf.PasswordRegex, err.Error())
		conn.Write([]byte("Login incorrect\r\n"))
		return
	}

	if !matched {
		conn.Write([]byte("Login incorrect\r\n"))
		return
	}

	// Session phase - authenticated
	uuidSession := uuid.New()
	sessionKey := "TELNET" + host + username

	tr.TraceEvent(tracer.Event{
		Msg:         "New TELNET Terminal Session",
		Protocol:    tracer.TELNET.String(),
		RemoteAddr:  conn.RemoteAddr().String(),
		SourceIp:    host,
		SourcePort:  port,
		Status:      tracer.Start.String(),
		ID:          uuidSession.String(),
		User:        username,
		Description: servConf.Description,
	})

	// Load history for LLM context
	var histories []plugins.Message
	if telnetStrategy.Sessions.HasKey(sessionKey) {
		histories = telnetStrategy.Sessions.Query(sessionKey)
	}

	// Interactive command loop
	for {
		// Display prompt (no newline - user types on same line)
		prompt := buildPrompt(username, servConf.ServerName)
		_, err := conn.Write([]byte(prompt))
		if err != nil {
			break
		}

		// Read command from user
		commandInput, err := readLine(conn)
		if err != nil {
			break
		}
		commandInput = strings.TrimSpace(commandInput)

		// Handle exit command
		if commandInput == "exit" {
			break
		}

		// Match command against regexes
		matched := false
		for _, command := range servConf.Commands {
			if command.Regex.MatchString(commandInput) {
				matched = true
				commandOutput := command.Handler

				// LLM integration - exactly like SSH
				if command.Plugin == plugins.LLMPluginName {
					llmProvider, err := plugins.FromStringToLLMProvider(servConf.Plugin.LLMProvider)
					if err != nil {
						log.Errorf("error: %s, fallback OpenAI", err.Error())
						llmProvider = plugins.OpenAI
					}
					llmHoneypot := plugins.BuildHoneypot(histories, tracer.TELNET, llmProvider, servConf)
					llmHoneypotInstance := plugins.InitLLMHoneypot(*llmHoneypot)
					if commandOutput, err = llmHoneypotInstance.ExecuteModel(commandInput); err != nil {
						log.Errorf("error ExecuteModel: %s, %s", commandInput, err.Error())
						commandOutput = "command not found"
					}
				}

				// Store command and response in history
				var newEntries []plugins.Message
				newEntries = append(newEntries, plugins.Message{Role: plugins.USER.String(), Content: commandInput})
				newEntries = append(newEntries, plugins.Message{Role: plugins.ASSISTANT.String(), Content: commandOutput})
				telnetStrategy.Sessions.Append(sessionKey, newEntries...)
				histories = append(histories, newEntries...)

				// Send response to client
				_, err := conn.Write([]byte(commandOutput + "\r\n"))
				if err != nil {
					break
				}

				// Trace interaction event
				tr.TraceEvent(tracer.Event{
					Msg:           "TELNET Terminal Session Interaction",
					RemoteAddr:    conn.RemoteAddr().String(),
					SourceIp:      host,
					SourcePort:    port,
					Status:        tracer.Interaction.String(),
					Command:       commandInput,
					CommandOutput: commandOutput,
					ID:            uuidSession.String(),
					Protocol:      tracer.TELNET.String(),
					Description:   servConf.Description,
					Handler:       command.Name,
				})

				break // Found match, exit command loop
			}
		}

		// If no command matched, send "command not found"
		if !matched {
			commandOutput := "command not found"
			_, err := conn.Write([]byte(commandOutput + "\r\n"))
			if err != nil {
				break
			}

			// Still trace the interaction even for unmatched commands
			tr.TraceEvent(tracer.Event{
				Msg:           "TELNET Terminal Session Interaction",
				RemoteAddr:    conn.RemoteAddr().String(),
				SourceIp:      host,
				SourcePort:    port,
				Status:        tracer.Interaction.String(),
				Command:       commandInput,
				CommandOutput: commandOutput,
				ID:            uuidSession.String(),
				Protocol:      tracer.TELNET.String(),
				Description:   servConf.Description,
				Handler:       "not_found",
			})
		}
	}

	// Trace session end
	tr.TraceEvent(tracer.Event{
		Msg:      "End TELNET Session",
		Status:   tracer.End.String(),
		ID:       uuidSession.String(),
		Protocol: tracer.TELNET.String(),
	})
}

func negotiateTelnet(conn net.Conn) {
	// Minimal telnet negotiation
	// Send what we support but don't wait for client response
	// Many simple telnet clients don't handle complex negotiation
	conn.Write([]byte{IAC, WILL, ECHO})
	conn.Write([]byte{IAC, WILL, SUPPRESS_GO_AHEAD})
	// Don't read or wait for responses
}

func drainInitialInput(conn net.Conn) {
	// Read and discard any IAC sequences the client sends back
	// Set a longer timeout to ensure we catch all negotiation bytes
	conn.SetReadDeadline(time.Now().Add(200 * time.Millisecond))

	buf := make([]byte, 256)
	for {
		_, err := conn.Read(buf)
		if err != nil {
			// Timeout or other error - we're done
			break
		}
	}

	// Reset to no deadline
	conn.SetReadDeadline(time.Time{})
}

func readLine(conn net.Conn) (string, error) {
	// Read one byte at a time until we get a newline
	var line []byte
	buf := make([]byte, 1)
	var i int = 0

	for {
		_, err := conn.Read(buf)
		if err != nil {
			return "", err
		}

		byte := buf[0]

		// Check for IAC (Interpret As Command) - skip it and next 2 bytes
		if byte == IAC && i+2 < 256 {
			// Read and discard the next 2 bytes
			conn.Read(buf) // command
			conn.Read(buf) // option
			continue
		}

		// Check for newline
		if byte == '\n' {
			break
		}

		// Skip carriage returns but keep other characters
		if byte != '\r' {
			line = append(line, byte)
		}

		i++
	}

	return string(line), nil
}

func stripIAC(data string) string {
	// Remove IAC sequences (IAC followed by two bytes)
	var result bytes.Buffer
	i := 0
	for i < len(data) {
		if data[i] == IAC && i+2 < len(data) {
			// Skip IAC sequence
			i += 3
		} else {
			result.WriteByte(data[i])
			i++
		}
	}
	return result.String()
}

func buildPrompt(user string, serverName string) string {
	return fmt.Sprintf("%s@%s:~$ ", user, serverName)
}
