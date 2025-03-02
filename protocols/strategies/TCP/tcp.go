package TCP

import (
	"fmt"
	"net"
	"time"

	"github.com/mariocandela/beelzebub/v3/parser"
	"github.com/mariocandela/beelzebub/v3/tracer"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type TCPStrategy struct {
}

func (tcpStrategy *TCPStrategy) Init(servConf parser.BeelzebubServiceConfiguration, tr tracer.Tracer) error {
	listen, err := net.Listen("tcp", servConf.Address)
	if err != nil {
		log.Errorf("Error during init TCP Protocol: %s", err.Error())
		return err
	}

	go func() {
		for {
			if conn, err := listen.Accept(); err == nil {
				go func() {
					conn.SetDeadline(time.Now().Add(time.Duration(servConf.DeadlineTimeoutSeconds) * time.Second))
					conn.Write(fmt.Appendf([]byte{}, "%s\n", servConf.Banner))

					buffer := make([]byte, 1024)
					command := ""

					if n, err := conn.Read(buffer); err == nil {
						command = string(buffer[:n])
					}

					host, port, _ := net.SplitHostPort(conn.RemoteAddr().String())

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
					conn.Close()
				}()
			}
		}
	}()

	log.WithFields(log.Fields{
		"port":   servConf.Address,
		"banner": servConf.Banner,
	}).Infof("Init service %s", servConf.Protocol)
	return nil
}
