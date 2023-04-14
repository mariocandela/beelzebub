package tracer

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Event struct {
	DateTime        string
	RemoteAddr      string
	Protocol        string
	Command         string
	CommandOutput   string
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

type Strategy func(event Event)

type Tracer interface {
	TraceEvent(event Event)
}

type tracer struct {
	strategy Strategy
}

var (
	eventsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "beelzebub",
		Name:      "events_total",
		Help:      "The total number of events",
	})
	eventsSSHTotal = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "beelzebub",
		Name:      "ssh_events_total",
		Help:      "The total number of SSH events",
	})
	eventsTCPTotal = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "beelzebub",
		Name:      "tcp_events_total",
		Help:      "The total number of TCP events",
	})
	eventsHTTPTotal = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "beelzebub",
		Name:      "http_events_total",
		Help:      "The total number of HTTP events",
	})
)

func Init(strategy Strategy) *tracer {
	return &tracer{
		strategy: strategy,
	}
}

func (tracer *tracer) TraceEvent(event Event) {
	event.DateTime = time.Now().UTC().Format(time.RFC3339)

	tracer.strategy(event)

	//Openmetrics
	eventsTotal.Inc()

	switch event.Protocol {
	case HTTP.String():
		eventsHTTPTotal.Inc()
		break
	case SSH.String():
		eventsSSHTotal.Inc()
		break
	case TCP.String():
		eventsTCPTotal.Inc()
		break
	}
}
