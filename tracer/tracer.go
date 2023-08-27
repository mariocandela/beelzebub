package tracer

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	log "github.com/sirupsen/logrus"
)

const Workers = 5

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

type (
	Protocol int
	Status   int
)

const (
	HTTP Protocol = iota
	SSH
	TCP
)

func (status Protocol) String() string {
	return [...]string{"HTTP", "SSH", "TCP"}[status]
}

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
	AddStrategy(strategy Strategy)
}

type tracer struct {
	strategies []Strategy
	eventsChan chan Event
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
	tracer := &tracer{
		strategies: make([]Strategy, 0, 20),
		eventsChan: make(chan Event, Workers),
	}

	tracer.AddStrategy(strategy)

	for i := 0; i < Workers; i++ {
		go func(i int) {
			log.Debug("Init trace worker: ", i)
			for event := range tracer.eventsChan {
				for _, strategy := range tracer.strategies {
					strategy(event)
				}
			}
		}(i)
	}

	return tracer
}

func (tracer *tracer) AddStrategy(strategy Strategy) {
	tracer.strategies = append(tracer.strategies, strategy)
}

func (tracer *tracer) TraceEvent(event Event) {
	event.DateTime = time.Now().UTC().Format(time.RFC3339)

	tracer.eventsChan <- event

	eventsTotal.Inc()

	switch event.Protocol {
	case HTTP.String():
		eventsHTTPTotal.Inc()
	case SSH.String():
		eventsSSHTotal.Inc()
	case TCP.String():
		eventsTCPTotal.Inc()
	}
}
