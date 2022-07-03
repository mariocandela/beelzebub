package tracer

import (
	"time"
)

type Strategy func(event Event)

type Tracer interface {
	TraceEvent(event Event)
}

type tracer struct {
	strategy Strategy
}

func Init(strategy Strategy) *tracer {
	return &tracer{
		strategy: strategy,
	}
}

func (tracer *tracer) TraceEvent(event Event) {
	event.DateTime = time.Now().UTC().Format(time.RFC3339)
	tracer.strategy(event)
}

type Event struct {
	DateTime        string
	RemoteAddr      string
	Protocol        string
	Command         string
	Status          string
	Msg             string
	ID              string
	Environ         string
	User            string
	Password        string
	Client          string
	Headers         string
	Cookies         string
	UserAgent       string
	HostHTTPRequest string
	Body            string
	HTTPMethod      string
	RequestURI      string
	Description     string
}

type Protocol int

const (
	HTTP Protocol = iota
	SSH
	TCP
)

func (status Protocol) String() string {
	return [...]string{"HTTP", "SSH", "TCP"}[status]
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
