package protocols

import (
	"beelzebub/parser"
	"beelzebub/tracer"
	log "github.com/sirupsen/logrus"
	"net"
	"time"
)

type TransmissionControlProtocolStrategy struct {
}

func (TCPStrategy *TransmissionControlProtocolStrategy) Init(beelzebubServiceConfiguration parser.BeelzebubServiceConfiguration, tr tracer.Tracer) error {
	listen, err := net.Listen("TCP", beelzebubServiceConfiguration.Address)
	if err != nil {
		log.Errorf("Error during init TCP Protocol: %s", err.Error())
		return err
	}
	defer listen.Close()

	go func() {
		for {
			if conn, err := listen.Accept(); err == nil {
				conn.SetDeadline(time.Now().Add(time.Duration(beelzebubServiceConfiguration.DeadlineTimeoutSeconds) * time.Second))
				go handleIncomingRequest(conn)
			}
		}
	}()

	log.WithFields(log.Fields{
		"port":   beelzebubServiceConfiguration.Address,
		"banner": beelzebubServiceConfiguration.Banner,
	}).Infof("Init service %s", beelzebubServiceConfiguration.Protocol)
	return nil
}

func handleIncomingRequest(conn net.Conn) {
	buffer := make([]byte, 1024)
	_, err := conn.Read(buffer)
	if err != nil {
		log.Fatal(err)
	}
	// respond
	conn.Write([]byte("Hi back!\n"))

	// close conn
	conn.Close()
}
