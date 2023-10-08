package strategies

import (
	"fmt"
	"github.com/mariocandela/beelzebub/v3/parser"
	"github.com/mariocandela/beelzebub/v3/plugins"
	"github.com/mariocandela/beelzebub/v3/tracer"
	"regexp"

	"strings"
	"time"

	"github.com/gliderlabs/ssh"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh/terminal"
)

type SSHStrategy struct {
}

func (sshStrategy *SSHStrategy) Init(beelzebubServiceConfiguration parser.BeelzebubServiceConfiguration, tr tracer.Tracer) error {
	go func() {
		server := &ssh.Server{
			Addr:        beelzebubServiceConfiguration.Address,
			MaxTimeout:  time.Duration(beelzebubServiceConfiguration.DeadlineTimeoutSeconds) * time.Second,
			IdleTimeout: time.Duration(beelzebubServiceConfiguration.DeadlineTimeoutSeconds) * time.Second,
			Version:     beelzebubServiceConfiguration.ServerVersion,
			Handler: func(sess ssh.Session) {
				uuidSession := uuid.New()

				tr.TraceEvent(tracer.Event{
					Msg:         "New SSH Session",
					Protocol:    tracer.SSH.String(),
					RemoteAddr:  sess.RemoteAddr().String(),
					Status:      tracer.Start.String(),
					ID:          uuidSession.String(),
					Environ:     strings.Join(sess.Environ(), ","),
					User:        sess.User(),
					Description: beelzebubServiceConfiguration.Description,
					Command:     sess.RawCommand(),
				})

				term := terminal.NewTerminal(sess, buildPrompt(sess.User(), beelzebubServiceConfiguration.ServerName))
				var histories []plugins.History
				for {
					commandInput, err := term.ReadLine()
					if err != nil {
						break
					}

					if commandInput == "exit" {
						break
					}
					for _, command := range beelzebubServiceConfiguration.Commands {
						matched, err := regexp.MatchString(command.Regex, commandInput)
						if err != nil {
							log.Errorf("Error regex: %s, %s", command.Regex, err.Error())
							continue
						}

						if matched {
							commandOutput := command.Handler

							if command.Plugin == plugins.ChatGPTPluginName {
								openAIGPTVirtualTerminal := plugins.Init(histories, beelzebubServiceConfiguration.Plugin.OpenAPIChatGPTSecretKey)

								if commandOutput, err = openAIGPTVirtualTerminal.GetCompletions(commandInput); err != nil {
									log.Errorf("Error GetCompletions: %s, %s", commandInput, err.Error())
									commandOutput = "command not found"
								}
							}

							histories = append(histories, plugins.History{Input: commandInput, Output: commandOutput})

							term.Write(append([]byte(commandOutput), '\n'))

							tr.TraceEvent(tracer.Event{
								Msg:           "New SSH Terminal Session",
								RemoteAddr:    sess.RemoteAddr().String(),
								Status:        tracer.Interaction.String(),
								Command:       commandInput,
								CommandOutput: commandOutput,
								ID:            uuidSession.String(),
								Protocol:      tracer.SSH.String(),
								Description:   beelzebubServiceConfiguration.Description,
							})
							break
						}
					}
				}
				tr.TraceEvent(tracer.Event{
					Msg:    "End SSH Session",
					Status: tracer.End.String(),
					ID:     uuidSession.String(),
				})
			},
			PasswordHandler: func(ctx ssh.Context, password string) bool {
				tr.TraceEvent(tracer.Event{
					Msg:         "New SSH attempt",
					Protocol:    tracer.SSH.String(),
					Status:      tracer.Stateless.String(),
					User:        ctx.User(),
					Password:    password,
					Client:      ctx.ClientVersion(),
					RemoteAddr:  ctx.RemoteAddr().String(),
					ID:          uuid.New().String(),
					Description: beelzebubServiceConfiguration.Description,
				})
				matched, err := regexp.MatchString(beelzebubServiceConfiguration.PasswordRegex, password)
				if err != nil {
					log.Errorf("Error regex: %s, %s", beelzebubServiceConfiguration.PasswordRegex, err.Error())
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
		"port":     beelzebubServiceConfiguration.Address,
		"commands": len(beelzebubServiceConfiguration.Commands),
	}).Infof("GetInstance service %s", beelzebubServiceConfiguration.Protocol)
	return nil
}

func buildPrompt(user string, serverName string) string {
	return fmt.Sprintf("%s@%s:~$ ", user, serverName)
}
