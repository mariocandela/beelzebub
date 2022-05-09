package tracer

import (
	log "github.com/sirupsen/logrus"
)

type Tracer struct {
}

func Init() *Tracer {
	return &Tracer{}
}

func (tracer *Tracer) TraceEvent(event Event) {
	log.WithFields(log.Fields{
		"status": event.Status.String(),
		"event":  event,
	}).Info("New Event")
}

type Event struct {
	RemoteAddr string
	Protocol   string
	Command    string
	Status     Status
	Msg        string
	ID         string
	Environ    string
	User       string
	Password   string
	Client     string
}

type Status int

const (
	Start Status = iota
	End
	Stateless
	Interaction
)

func (status Status) String() string {
	return [...]string{"Start", "End", "Stateless", "Interaction"}[status]
}
