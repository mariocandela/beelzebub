package tracer

import (
	log "github.com/sirupsen/logrus"
	"net/http"
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
	RemoteAddr      string
	Protocol        Protocol
	Command         string
	Status          Status
	Msg             string
	ID              string
	Environ         string
	User            string
	Password        string
	Client          string
	Headers         http.Header
	Cookies         []*http.Cookie
	UserAgent       string
	HostHTTPRequest string
	Body            string
	HTTPMethod      string
	RequestURI      string
}

type Protocol int

const (
	HTTP Protocol = iota
	SSH
)

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
