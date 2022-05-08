package protocols

import (
	"beelzebub/parser"
	"fmt"
	"github.com/gliderlabs/ssh"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh/terminal"
	"regexp"
	"time"
)

type SecureShellStrategy struct {
}

func (SSHStrategy *SecureShellStrategy) Init(beelzebubServiceConfiguration parser.BeelzebubServiceConfiguration) error {
	go func() {
		server := &ssh.Server{
			Addr:        beelzebubServiceConfiguration.Address,
			MaxTimeout:  time.Duration(beelzebubServiceConfiguration.DeadlineTimeoutSeconds) * time.Second,
			IdleTimeout: time.Duration(beelzebubServiceConfiguration.DeadlineTimeoutSeconds) * time.Second,
			Version:     beelzebubServiceConfiguration.ServerVersion,
			Handler: func(sess ssh.Session) {
				uuidSession := uuid.New()
				traceSessionStart(sess, uuidSession)
				term := terminal.NewTerminal(sess, buildPrompt(sess.User(), beelzebubServiceConfiguration.ServerName))
				for {
					commandInput, err := term.ReadLine()
					if err != nil {
						break
					}
					traceCommand(commandInput, uuidSession)
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
							term.Write(append([]byte(command.Handler), '\n'))
							break
						}
					}
				}
				traceSessionEnd(sess, uuidSession)
			},
			PasswordHandler: func(ctx ssh.Context, password string) bool {
				traceAttempt(ctx, password)
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
	}).Infof("Init service %s", beelzebubServiceConfiguration.Protocol)
	return nil
}

func buildPrompt(user string, serverName string) string {
	return fmt.Sprintf("%s@%s:~$ ", user, serverName)
}

func traceAttempt(ctx ssh.Context, password string) {
	log.WithFields(log.Fields{
		"remoteAddr": ctx.RemoteAddr(),
		"user":       ctx.User(),
		"password":   password,
		"client":     ctx.ClientVersion(),
	}).Info("New SSH attempt")
}

func traceSessionStart(sess ssh.Session, uuidSession uuid.UUID) {
	log.WithFields(log.Fields{
		"uuidSession": uuidSession,
		"remoteAddr":  sess.RemoteAddr(),
		"command":     sess.Command(),
		"environ":     sess.Environ(),
		"user":        sess.User(),
	}).Info("New SSH Session")
}

func traceSessionEnd(sess ssh.Session, uuidSession uuid.UUID) {
	log.WithFields(log.Fields{
		"uuidSession": uuidSession,
		"remoteAddr":  sess.RemoteAddr(),
		"command":     sess.Command(),
		"environ":     sess.Environ(),
		"user":        sess.User(),
	}).Info("End SSH Session")
}

func traceCommand(command string, uuidSession uuid.UUID) {
	log.WithFields(log.Fields{
		"uuidSession": uuidSession,
		"command":     command,
	}).Info("New SSH Command")
}
