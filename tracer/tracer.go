package tracer

import (
	"net/http"
)

type Strategy func(event Event)

type Tracer struct {
	strategy Strategy
}

func Init(strategy Strategy) *Tracer {
	return &Tracer{
		strategy: strategy,
	}
}

func (tracer *Tracer) TraceEvent(event Event) {
	tracer.strategy(event)
}

type Event struct {
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

func (status Protocol) String() string {
	return [...]string{"HTTP", "SSH"}[status]
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
