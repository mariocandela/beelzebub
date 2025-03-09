package SSH

import (
	"fmt"
	"net"
	"regexp"
	"strings"
	"time"

	"github.com/mariocandela/beelzebub/v3/parser"
	"github.com/mariocandela/beelzebub/v3/plugins"
	"github.com/mariocandela/beelzebub/v3/tracer"

	"github.com/gliderlabs/ssh"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"golang.org/x/term"
)

type SSHStrategy struct {
	Sessions map[string][]plugins.Message
}

func (sshStrategy *SSHStrategy) Init(servConf parser.BeelzebubServiceConfiguration, tr tracer.Tracer) error {
	if sshStrategy.Sessions == nil {
		sshStrategy.Sessions = make(map[string][]plugins.Message)
	}
	go func() {
		server := &ssh.Server{
			Addr:        servConf.Address,
			MaxTimeout:  time.Duration(servConf.DeadlineTimeoutSeconds) * time.Second,
			IdleTimeout: time.Duration(servConf.DeadlineTimeoutSeconds) * time.Second,
			Version:     servConf.ServerVersion,
			Handler: func(sess ssh.Session) {
				uuidSession := uuid.New()

				host, port, _ := net.SplitHostPort(sess.RemoteAddr().String())
				sessionKey := host + sess.User()

				// Inline SSH command
				if sess.RawCommand() != "" {
					for _, command := range servConf.Commands {
						matched, err := regexp.MatchString(command.Regex, sess.RawCommand())
						if err != nil {
							log.Errorf("Error regex: %s, %s", command.Regex, err.Error())
							continue
						}

						if matched {
							commandOutput := command.Handler

							if command.Plugin == plugins.LLMPluginName {

								llmProvider, err := plugins.FromStringToLLMProvider(servConf.Plugin.LLMProvider)

								if err != nil {
									log.Errorf("Error: %s", err.Error())
									commandOutput = "command not found"
									llmProvider = plugins.OpenAI
								}

								var histories []plugins.Message
								if sshStrategy.Sessions[sessionKey] != nil {
									histories = sshStrategy.Sessions[sessionKey]
								}
								llmHoneypot := plugins.LLMHoneypot{
									Histories:    histories,
									OpenAIKey:    servConf.Plugin.OpenAISecretKey,
									Protocol:     tracer.SSH,
									Host:         servConf.Plugin.Host,
									Model:        servConf.Plugin.LLMModel,
									Provider:     llmProvider,
									CustomPrompt: servConf.Plugin.Prompt,
								}

								llmHoneypotInstance := plugins.InitLLMHoneypot(llmHoneypot)

								if commandOutput, err = llmHoneypotInstance.ExecuteModel(sess.RawCommand()); err != nil {
									log.Errorf("Error ExecuteModel: %s, %s", sess.RawCommand(), err.Error())
									commandOutput = "command not found"
								}
							}

							sess.Write(append([]byte(commandOutput), '\n'))

							tr.TraceEvent(tracer.Event{
								Msg:           "New SSH Session",
								Protocol:      tracer.SSH.String(),
								RemoteAddr:    sess.RemoteAddr().String(),
								SourceIp:      host,
								SourcePort:    port,
								Status:        tracer.Start.String(),
								ID:            uuidSession.String(),
								Environ:       strings.Join(sess.Environ(), ","),
								User:          sess.User(),
								Description:   servConf.Description,
								Command:       sess.RawCommand(),
								CommandOutput: commandOutput,
							})
							var histories []plugins.Message
							histories = append(histories, plugins.Message{Role: plugins.USER.String(), Content: sess.RawCommand()})
							histories = append(histories, plugins.Message{Role: plugins.ASSISTANT.String(), Content: commandOutput})
							sshStrategy.Sessions[sessionKey] = histories
							tr.TraceEvent(tracer.Event{
								Msg:    "End SSH Session",
								Status: tracer.End.String(),
								ID:     uuidSession.String(),
							})
							return
						}
					}
				}

				tr.TraceEvent(tracer.Event{
					Msg:         "New SSH Session",
					Protocol:    tracer.SSH.String(),
					RemoteAddr:  sess.RemoteAddr().String(),
					SourceIp:    host,
					SourcePort:  port,
					Status:      tracer.Start.String(),
					ID:          uuidSession.String(),
					Environ:     strings.Join(sess.Environ(), ","),
					User:        sess.User(),
					Description: servConf.Description,
				})

				terminal := term.NewTerminal(sess, buildPrompt(sess.User(), servConf.ServerName))
				var histories []plugins.Message
				if sshStrategy.Sessions[sessionKey] != nil {
					histories = sshStrategy.Sessions[sessionKey]
				}

				for {
					commandInput, err := terminal.ReadLine()
					if err != nil {
						break
					}

					if commandInput == "exit" {
						break
					}
					for _, command := range servConf.Commands {
						matched, err := regexp.MatchString(command.Regex, commandInput)
						if err != nil {
							log.Errorf("Error regex: %s, %s", command.Regex, err.Error())
							continue
						}

						if matched {
							commandOutput := command.Handler

							if command.Plugin == plugins.LLMPluginName {

								llmProvider, err := plugins.FromStringToLLMProvider(servConf.Plugin.LLMProvider)

								if err != nil {
									log.Errorf("Error: %s, fallback OpenAI", err.Error())
									llmProvider = plugins.OpenAI
								}

								llmHoneypot := plugins.LLMHoneypot{
									Histories:    histories,
									OpenAIKey:    servConf.Plugin.OpenAISecretKey,
									Protocol:     tracer.SSH,
									Host:         servConf.Plugin.Host,
									Model:        servConf.Plugin.LLMModel,
									Provider:     llmProvider,
									CustomPrompt: servConf.Plugin.Prompt,
								}

								llmHoneypotInstance := plugins.InitLLMHoneypot(llmHoneypot)

								if commandOutput, err = llmHoneypotInstance.ExecuteModel(commandInput); err != nil {
									log.Errorf("Error ExecuteModel: %s, %s", commandInput, err.Error())
									commandOutput = "command not found"
								}
							}

							histories = append(histories, plugins.Message{Role: plugins.USER.String(), Content: commandInput})
							histories = append(histories, plugins.Message{Role: plugins.ASSISTANT.String(), Content: commandOutput})

							terminal.Write(append([]byte(commandOutput), '\n'))

							tr.TraceEvent(tracer.Event{
								Msg:           "New SSH Terminal Session",
								RemoteAddr:    sess.RemoteAddr().String(),
								SourceIp:      host,
								SourcePort:    port,
								Status:        tracer.Interaction.String(),
								Command:       commandInput,
								CommandOutput: commandOutput,
								ID:            uuidSession.String(),
								Protocol:      tracer.SSH.String(),
								Description:   servConf.Description,
							})
							break
						}
					}
				}
				sshStrategy.Sessions[sessionKey] = histories
				tr.TraceEvent(tracer.Event{
					Msg:    "End SSH Session",
					Status: tracer.End.String(),
					ID:     uuidSession.String(),
				})
			},
			PasswordHandler: func(ctx ssh.Context, password string) bool {
				host, port, _ := net.SplitHostPort(ctx.RemoteAddr().String())

				tr.TraceEvent(tracer.Event{
					Msg:         "New SSH attempt",
					Protocol:    tracer.SSH.String(),
					Status:      tracer.Stateless.String(),
					User:        ctx.User(),
					Password:    password,
					Client:      ctx.ClientVersion(),
					RemoteAddr:  ctx.RemoteAddr().String(),
					SourceIp:    host,
					SourcePort:  port,
					ID:          uuid.New().String(),
					Description: servConf.Description,
				})
				matched, err := regexp.MatchString(servConf.PasswordRegex, password)
				if err != nil {
					log.Errorf("Error regex: %s, %s", servConf.PasswordRegex, err.Error())
					return false
				}
				return matched
			},
		}
		err := server.ListenAndServe()
		if err != nil {
			log.Errorf("Error during init SSH Protocol: %s", err.Error())
		}
	}()

	log.WithFields(log.Fields{
		"port":     servConf.Address,
		"commands": len(servConf.Commands),
	}).Infof("GetInstance service %s", servConf.Protocol)
	return nil
}

func buildPrompt(user string, serverName string) string {
	return fmt.Sprintf("%s@%s:~$ ", user, serverName)
}
