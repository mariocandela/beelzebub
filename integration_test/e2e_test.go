package integration_test

import (
	"beelzebub/builder"
	"beelzebub/parser"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/suite"
	"net/http"
	"testing"
)

type e2eTestSuite struct {
	suite.Suite
}

func TestE2ETestSuite(t *testing.T) {
	suite.Run(t, &e2eTestSuite{})
}

func (s *e2eTestSuite) SetupSuite() {
	beelzebubConfigPath := "./configurations/beelzebub.yaml"
	servicesConfigDirectory := "./configurations/services/"

	parser := parser.Init(beelzebubConfigPath, servicesConfigDirectory)

	coreConfigurations, err := parser.ReadConfigurationsCore()
	s.Require().NoError(err)

	beelzebubServicesConfiguration, err := parser.ReadConfigurationsServices()
	s.Require().NoError(err)

	beelzebubBuilder := builder.NewBuilder()

	director := builder.NewDirector(beelzebubBuilder)

	beelzebubBuilder, err = director.BuildBeelzebub(coreConfigurations, beelzebubServicesConfiguration)
	s.Require().NoError(err)
	defer beelzebubBuilder.Close()

	beelzebubReady := make(chan bool)

	s.Require().NoError(beelzebubBuilder.Run())

	<-beelzebubReady
}

func (s *e2eTestSuite) Test_EndToEnd_InvokeHTTPHoneypot() {
	response, err := resty.New().R().
		SetHeader("Content-Type", "application/json").
		Get("http://localhost/index.php")

	s.NoError(err)
	s.Equal(http.StatusOK, response.StatusCode)
	s.Equal("mocked response", string(response.Body()))

	response, err = resty.New().R().
		SetHeader("Content-Type", "application/json").
		Get("http://localhost/wp-admin.php")

	s.NoError(err)
	s.Equal(http.StatusBadRequest, response.StatusCode)
	s.Equal("mocked response", string(response.Body()))
}
