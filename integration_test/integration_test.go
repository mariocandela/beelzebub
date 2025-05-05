package integration

import (
	"encoding/json"
	"github.com/mariocandela/beelzebub/v3/builder"
	"github.com/mariocandela/beelzebub/v3/parser"
	"github.com/mariocandela/beelzebub/v3/tracer"
	"net"
	"net/http"
	"os"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/melbahja/goph"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/ssh"
)

type IntegrationTestSuite struct {
	suite.Suite
	beelzebubBuilder *builder.Builder
	prometheusHost   string
	httpHoneypotHost string
	tcpHoneypotHost  string
	sshHoneypotHost  string
	rabbitMQURI      string
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (suite *IntegrationTestSuite) SetupSuite() {
	suite.T().Helper()
	if os.Getenv("INTEGRATION") == "" {
		suite.T().Skip("skipping integration tests, set environment variable INTEGRATION")
	}
	suite.httpHoneypotHost = "http://localhost:8080"
	suite.tcpHoneypotHost = "localhost:3306"
	suite.sshHoneypotHost = "localhost"
	suite.prometheusHost = "http://localhost:2112/metrics"

	beelzebubConfigPath := "./configurations/beelzebub.yaml"
	servicesConfigDirectory := "./configurations/services/"

	parser := parser.Init(beelzebubConfigPath, servicesConfigDirectory)

	coreConfigurations, err := parser.ReadConfigurationsCore()
	suite.Require().NoError(err)
	suite.rabbitMQURI = coreConfigurations.Core.Tracings.RabbitMQ.URI

	beelzebubServicesConfiguration, err := parser.ReadConfigurationsServices()
	suite.Require().NoError(err)

	suite.beelzebubBuilder = builder.NewBuilder()

	director := builder.NewDirector(suite.beelzebubBuilder)

	suite.beelzebubBuilder, err = director.BuildBeelzebub(coreConfigurations, beelzebubServicesConfiguration)
	suite.Require().NoError(err)

	suite.Require().NoError(suite.beelzebubBuilder.Run())
}

func (suite *IntegrationTestSuite) TestInvokeHTTPHoneypot() {
	response, err := resty.New().R().
		Get(suite.httpHoneypotHost + "/index.php")

	response.Header().Del("Date")

	suite.Require().NoError(err)
	suite.Equal(http.StatusOK, response.StatusCode())
	suite.Equal(http.Header{"Content-Length": []string{"15"}, "Content-Type": []string{"text/html"}, "Server": []string{"Apache/2.4.53 (Debian)"}, "X-Powered-By": []string{"PHP/7.4.29"}}, response.Header())
	suite.Equal("mocked response", string(response.Body()))

	response, err = resty.New().R().
		Get(suite.httpHoneypotHost + "/wp-admin")

	suite.Require().NoError(err)
	suite.Equal(http.StatusBadRequest, response.StatusCode())
	suite.Equal("mocked response", string(response.Body()))
}

func (suite *IntegrationTestSuite) TestInvokeTCPHoneypot() {
	tcpAddr, err := net.ResolveTCPAddr("tcp", suite.tcpHoneypotHost)
	suite.Require().NoError(err)

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	suite.Require().NoError(err)
	defer conn.Close()

	_, err = conn.Write([]byte("hello!"))
	suite.Require().NoError(err)

	reply := make([]byte, 1024)

	n, err := conn.Read(reply)
	suite.Require().NoError(err)

	suite.Equal("8.0.29\n", string(reply[:n]))
}

func (suite *IntegrationTestSuite) TestInvokeSSHHoneypot() {
	client, err := goph.NewConn(
		&goph.Config{
			User:     "root",
			Addr:     suite.sshHoneypotHost,
			Port:     2222,
			Auth:     goph.Password("root"),
			Callback: ssh.InsecureIgnoreHostKey(),
		})
	suite.Require().NoError(err)
	defer client.Close()

	out, err := client.Run("")
	suite.Require().NoError(err)

	suite.Equal("root@ubuntu:~$ ", string(out))
}

func (suite *IntegrationTestSuite) TestRabbitMQ() {
	conn, err := amqp.Dial(suite.rabbitMQURI)
	suite.Require().NoError(err)
	defer conn.Close()

	ch, err := conn.Channel()
	suite.Require().NoError(err)
	defer ch.Close()

	msgs, err := ch.Consume("event", "", true, false, false, false, nil)
	suite.Require().NoError(err)

	//Invoke HTTP Honeypot
	response, err := resty.New().R().Get(suite.httpHoneypotHost + "/index.php")

	suite.Require().NoError(err)
	suite.Equal(http.StatusOK, response.StatusCode())

	for msg := range msgs {
		var event tracer.Event
		err := json.Unmarshal(msg.Body, &event)
		suite.Require().NoError(err)

		suite.Equal("GET", event.HTTPMethod)
		suite.Equal("/index.php", event.RequestURI)
		break
	}

}
func (suite *IntegrationTestSuite) TestPrometheus() {
	//Invoke HTTP Honeypot
	response, err := resty.New().R().Get(suite.httpHoneypotHost + "/index.php")

	suite.Require().NoError(err)
	suite.Equal(http.StatusOK, response.StatusCode())

	response, err = resty.New().R().Get(suite.prometheusHost)

	suite.Require().NoError(err)
	suite.Equal(http.StatusOK, response.StatusCode())
}

func (suite *IntegrationTestSuite) TestShutdownBeelzebub() {
	suite.Require().NoError(suite.beelzebubBuilder.Close())
}
