package integration

import (
	"beelzebub/builder"
	"beelzebub/parser"
	"github.com/go-resty/resty/v2"
	"github.com/melbahja/goph"
	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/ssh"
	"net"
	"net/http"
	"os"
	"testing"
)

type IntegrationTestSuite struct {
	suite.Suite
	beelzebubBuilder *builder.Builder
	httpHoneypotHost string
	tcpHoneypotHost  string
	sshHoneypotHost  string
}

func (suite *IntegrationTestSuite) skipIntegration() {
	suite.T().Helper()
	if os.Getenv("INTEGRATION") == "" {
		suite.T().Skip("skipping integration tests, set environment variable INTEGRATION")
	}
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (suite *IntegrationTestSuite) SetupSuite() {
	suite.skipIntegration()

	suite.httpHoneypotHost = "http://127.0.0.1:8080"
	suite.tcpHoneypotHost = "127.0.0.1:3306"
	suite.sshHoneypotHost = "127.0.0.1"

	beelzebubConfigPath := "./configurations/beelzebub.yaml"
	servicesConfigDirectory := "./configurations/services/"

	parser := parser.Init(beelzebubConfigPath, servicesConfigDirectory)

	coreConfigurations, err := parser.ReadConfigurationsCore()
	suite.Require().NoError(err)

	beelzebubServicesConfiguration, err := parser.ReadConfigurationsServices()
	suite.Require().NoError(err)

	suite.beelzebubBuilder = builder.NewBuilder()

	director := builder.NewDirector(suite.beelzebubBuilder)

	suite.beelzebubBuilder, err = director.BuildBeelzebub(coreConfigurations, beelzebubServicesConfiguration)
	suite.Require().NoError(err)

	suite.Require().NoError(suite.beelzebubBuilder.Run())
}

func (suite *IntegrationTestSuite) TestEndToEndInvokeHTTPHoneypot() {
	response, err := resty.New().R().
		Get(suite.httpHoneypotHost + "/index.php")

	suite.Require().NoError(err)
	suite.Equal(http.StatusOK, response.StatusCode())
	suite.Equal("mocked response", string(response.Body()))

	response, err = resty.New().R().
		Get(suite.httpHoneypotHost + "/wp-admin")

	suite.Require().NoError(err)
	suite.Equal(http.StatusBadRequest, response.StatusCode())
	suite.Equal("mocked response", string(response.Body()))
}

func (suite *IntegrationTestSuite) TestEndToEndInvokeTCPHoneypot() {
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

func (suite *IntegrationTestSuite) TestEndToEndInvokeSSHHoneypot() {
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

//TODO test rabbitmq

func (suite *IntegrationTestSuite) TestShutdownBeelzebub() {
	suite.Require().NoError(suite.beelzebubBuilder.Close())
}
