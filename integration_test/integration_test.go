package integration

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"github.com/go-resty/resty/v2"
	"github.com/mariocandela/beelzebub/v3/builder"
	"github.com/mariocandela/beelzebub/v3/parser"
	"github.com/mariocandela/beelzebub/v3/tracer"
	"github.com/melbahja/goph"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/ssh"
	"math/big"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

type IntegrationTestSuite struct {
	suite.Suite
	beelzebubBuilder *builder.Builder
	prometheusHost   string
	httpHoneypotHost string
	tcpHoneypotHost  string
	sshHoneypotHost  string
	rabbitMQURI      string

	tlsCertPath string
	tlsKeyPath  string
	tlsCleanup  func()
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func generateTLSCertInCertsDir(t *testing.T) (certPath, keyPath string, cleanup func()) {
	t.Helper()

	certsDir := filepath.Join(".", "certs")
	if err := os.MkdirAll(certsDir, 0755); err != nil {
		t.Fatalf("failed to create certs directory: %v", err)
	}

	certPath = filepath.Join(certsDir, "tls.crt")
	keyPath = filepath.Join(certsDir, "tls.key")

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate rsa key: %v", err)
	}

	tmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(time.Now().UnixNano()),
		Subject:               pkix.Name{Organization: []string{"IntegrationTest"}},
		NotBefore:             time.Now().Add(-1 * time.Minute),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:              []string{"localhost"},
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &priv.PublicKey, priv)
	if err != nil {
		t.Fatalf("failed to create certificate: %v", err)
	}

	certFile, err := os.Create(certPath)
	if err != nil {
		t.Fatalf("failed to create cert file: %v", err)
	}
	if err := pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		_ = certFile.Close()
		t.Fatalf("failed to encode cert PEM: %v", err)
	}
	if err := certFile.Close(); err != nil {
		t.Fatalf("failed to close cert file: %v", err)
	}

	keyFile, err := os.Create(keyPath)
	if err != nil {
		t.Fatalf("failed to create key file: %v", err)
	}
	if err := pem.Encode(keyFile, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)}); err != nil {
		_ = keyFile.Close()
		t.Fatalf("failed to encode key PEM: %v", err)
	}
	if err := keyFile.Close(); err != nil {
		t.Fatalf("failed to close key file: %v", err)
	}

	cleanup = func() {
		_ = os.Remove(certPath)
		_ = os.Remove(keyPath)
	}

	return certPath, keyPath, cleanup
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

	suite.tlsCertPath, suite.tlsKeyPath, suite.tlsCleanup = generateTLSCertInCertsDir(suite.T())

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

func (suite *IntegrationTestSuite) TestInvokeTLSHoneypot() {
	type eventCollector struct {
		mu     sync.Mutex
		events []tracer.Event
	}
	var collector eventCollector

	// Store original tracer singleton strategy
	tr := tracer.GetInstance(nil)
	originalStrategy := tr.GetStrategy()

	wrapperStrategy := func(event tracer.Event) {
		// Wrap original tracer strategy to not lose functionalities
		if originalStrategy != nil {
			originalStrategy(event)
		}

		// fetch last event, it will be used to check for TLSServerName
		collector.mu.Lock()
		collector.events = append(collector.events, event)
		collector.mu.Unlock()
	}

	// update strategy with wrapper, set original strategy at function exit
	tr.SetStrategy(wrapperStrategy)
	defer tr.SetStrategy(originalStrategy)

	client := resty.New()
	client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})

	response, err := client.R().Get("https://localhost:8443/secure-index.php")
	suite.Require().NoError(err)
	suite.Equal(http.StatusOK, response.StatusCode())
	suite.Equal("mocked secure response", string(response.Body()))
	headers := response.Header()
	suite.Equal("text/html", strings.TrimSpace(headers.Get("Content-Type")))
	suite.Equal("Apache/2.4.53 (Debian Secure)", strings.TrimSpace(headers.Get("Server")))
	suite.Equal("PHP/7.4.29 Secure", strings.TrimSpace(headers.Get("X-Powered-By")))

	response, err = client.R().Get("https://localhost:8443/secure-wp-admin")
	suite.Require().NoError(err)
	suite.Equal(http.StatusBadRequest, response.StatusCode())
	suite.Equal("mocked secure response", string(response.Body()))

	response, err = client.R().Get("https://localhost:8443/unknown-path")
	suite.Require().NoError(err)
	suite.Equal(http.StatusNotFound, response.StatusCode())
	suite.Equal("Secure not found!", string(response.Body()))

	// throttle cpu to wait event processing
	time.Sleep(100 * time.Millisecond)

	collector.mu.Lock()
	defer collector.mu.Unlock()
	found := false
	for _, ev := range collector.events {
		if ev.TLSServerName == "localhost" {
			found = true
			break
		}
	}
	suite.True(found, "Expected to find event with TLSServerName 'localhost'")
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

func (suite *IntegrationTestSuite) TearDownSuite() {
	suite.Require().NoError(suite.beelzebubBuilder.Close())
	if suite.tlsCleanup != nil {
		suite.tlsCleanup()
	}
}
