package SSH

import (
	"fmt"
	"net"
	"regexp"
	"strings"
	"time"

	"github.com/mariocandela/beelzebub/v3/historystore"
	"github.com/mariocandela/beelzebub/v3/parser"
	"github.com/mariocandela/beelzebub/v3/plugins"
	"github.com/mariocandela/beelzebub/v3/tracer"

	"github.com/gliderlabs/ssh"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"golang.org/x/term"
)

// If REPLAYFROMCACHE is true, then cached responses from previous LLM results will be used.
// TODO(bryannolen): Add this to Service Configuration (or maybe per command?)
const REPLAYFROMCACHE = true

type SSHStrategy struct {
	Sessions *historystore.HistoryStore
}

func (sshStrategy *SSHStrategy) Init(servConf parser.BeelzebubServiceConfiguration, tr tracer.Tracer) error {
	if sshStrategy.Sessions == nil {
		sshStrategy.Sessions = historystore.NewHistoryStore()
	}
	go sshStrategy.Sessions.HistoryCleaner()
	go func() {
		server := &ssh.Server{
			Addr:        servConf.Address,
			MaxTimeout:  time.Duration(servConf.DeadlineTimeoutSeconds) * time.Second,
			IdleTimeout: time.Duration(servConf.DeadlineTimeoutSeconds) * time.Second,
			Version:     servConf.ServerVersion,
			Handler: func(sess ssh.Session) {
				uuidSession := uuid.New()

				host, port, _ := net.SplitHostPort(sess.RemoteAddr().String())
				sessionKey := "SSH" + host + sess.User()

				// Inline SSH command
				if sess.RawCommand() != "" {
					var histories []plugins.Message
					inMsg := plugins.Message{Role: plugins.USER.String(), Content: sess.RawCommand()}
					commandOutput := "command not found"
					haveCachedAnswer := false
					if sshStrategy.Sessions.HasKey(sessionKey) {
						histories = sshStrategy.Sessions.Query(sessionKey)
						if REPLAYFROMCACHE {
							cacheReply := sshStrategy.Sessions.QueryConversations(sessionKey, inMsg)
							if cacheReply != nil {
								commandOutput = cacheReply.Output.Content
								haveCachedAnswer = true
							}
						}
					}

					for _, command := range servConf.Commands {
						if command.Regex.MatchString(sess.RawCommand()) {
							if !haveCachedAnswer {
								commandOutput = command.Handler
								if command.Plugin == plugins.LLMPluginName {
									llmProvider, err := plugins.FromStringToLLMProvider(servConf.Plugin.LLMProvider)
									if err != nil {
										log.Errorf("error: %s, fallback OpenAI", err.Error())
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
									if commandOutput, err = llmHoneypotInstance.ExecuteModel(sess.RawCommand()); err != nil {
										log.Errorf("error ExecuteModel: %s, %s", sess.RawCommand(), err.Error())
										commandOutput = "command not found"
									}
								}
							}
							// Append the new entries to the store.
							outMsg := plugins.Message{Role: plugins.ASSISTANT.String(), Content: commandOutput}
							sshStrategy.Sessions.Append(sessionKey, inMsg, outMsg)
							sshStrategy.Sessions.AppendConverstion(sessionKey, historystore.Conversation{Input: inMsg, Output: outMsg})

							sess.Write(append([]byte(commandOutput), '\n'))
							handlerName := command.Name
							if haveCachedAnswer {
								handlerName = handlerName + " cached"
							}
							tr.TraceEvent(tracer.Event{
								Msg:           "SSH Raw Command",
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
								Handler:       handlerName,
							})
							return
						}
					}
				} // end raw command handler.

				tr.TraceEvent(tracer.Event{
					Msg:         "New SSH Terminal Session",
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
				if sshStrategy.Sessions.HasKey(sessionKey) {
					// Load from the main HistoryStore only once at the start of the interactive session.
					histories = sshStrategy.Sessions.Query(sessionKey)
				}

				for {
					commandInput, err := terminal.ReadLine()
					inMsg := plugins.Message{Role: plugins.USER.String(), Content: commandInput}
					commandOutput := "command not found"
					if err != nil {
						break
					}
					if commandInput == "exit" {
						break
					}
					haveCachedAnswer := false
					if REPLAYFROMCACHE {
						cacheReply := sshStrategy.Sessions.QueryConversations(sessionKey, inMsg)
						if cacheReply != nil {
							commandOutput = cacheReply.Output.Content
							haveCachedAnswer = true
						}
					}
					for _, command := range servConf.Commands {
						if command.Regex.MatchString(commandInput) {
							if !haveCachedAnswer {
								commandOutput = command.Handler
								if command.Plugin == plugins.LLMPluginName {
									llmProvider, err := plugins.FromStringToLLMProvider(servConf.Plugin.LLMProvider)
									if err != nil {
										log.Errorf("error: %s, fallback OpenAI", err.Error())
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
										log.Errorf("error ExecuteModel: %s, %s", commandInput, err.Error())
										commandOutput = "command not found"
									}
								}
							}
							outMsg := plugins.Message{Role: plugins.ASSISTANT.String(), Content: commandOutput}
							// Stash the new entries to the HistoryStore, and update the local history used for this running session.
							sshStrategy.Sessions.Append(sessionKey, inMsg, outMsg)
							sshStrategy.Sessions.AppendConverstion(sessionKey, historystore.Conversation{Input: inMsg, Output: outMsg})
							histories = append(histories, inMsg, outMsg)

							terminal.Write(append([]byte(commandOutput), '\n'))

							handlerName := command.Name
							if haveCachedAnswer {
								handlerName = handlerName + " cached"
							}

							tr.TraceEvent(tracer.Event{
								Msg:           "SSH Terminal Session Interaction",
								RemoteAddr:    sess.RemoteAddr().String(),
								SourceIp:      host,
								SourcePort:    port,
								Status:        tracer.Interaction.String(),
								Command:       commandInput,
								CommandOutput: commandOutput,
								ID:            uuidSession.String(),
								Protocol:      tracer.SSH.String(),
								Description:   servConf.Description,
								Handler:       handlerName,
							})
							break // Inner range over commands.
						}
					}
				}

				tr.TraceEvent(tracer.Event{
					Msg:      "End SSH Session",
					Status:   tracer.End.String(),
					ID:       uuidSession.String(),
					Protocol: tracer.SSH.String(),
				})
			},
			PasswordHandler: func(ctx ssh.Context, password string) bool {
				host, port, _ := net.SplitHostPort(ctx.RemoteAddr().String())

				tr.TraceEvent(tracer.Event{
					Msg:         "New SSH Login Attempt",
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
					log.Errorf("error regex: %s, %s", servConf.PasswordRegex, err.Error())
					return false
				}
				return matched
			},
		}
		err := server.ListenAndServe()
		if err != nil {
			log.Errorf("error during init SSH Protocol: %s", err.Error())
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
