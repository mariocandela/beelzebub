package strategies

import (
	"fmt"
	"github.com/mariocandela/beelzebub/v3/parser"
	"github.com/mariocandela/beelzebub/v3/tracer"
	"net"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type TCPStrategy struct {
}

func (tcpStrategy *TCPStrategy) Init(beelzebubServiceConfiguration parser.BeelzebubServiceConfiguration, tr tracer.Tracer) error {
	listen, err := net.Listen("tcp", beelzebubServiceConfiguration.Address)
	if err != nil {
		log.Errorf("Error during init TCP Protocol: %s", err.Error())
		return err
	}

	go func() {
		for {
			if conn, err := listen.Accept(); err == nil {
				go func() {
					conn.SetDeadline(time.Now().Add(time.Duration(beelzebubServiceConfiguration.DeadlineTimeoutSeconds) * time.Second))
					conn.Write([]byte(fmt.Sprintf("%s\n", beelzebubServiceConfiguration.Banner)))

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
						ID:          uuid.New().String(),
						Description: beelzebubServiceConfiguration.Description,
					})
					conn.Close()
				}()
			}
		}
	}()

	log.WithFields(log.Fields{
		"port":   beelzebubServiceConfiguration.Address,
		"banner": beelzebubServiceConfiguration.Banner,
	}).Infof("Init service %s", beelzebubServiceConfiguration.Protocol)
	return nil
}
