package parser

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func mockReadfilebytesConfigurationsCore(filePath string) ([]byte, error) {
	configurationsCoreBytes := []byte(`
core:
  logging:
    debug: false
    debugReportCaller: false
    logDisableTimestamp: true
    logsPath: ./logs
  tracings:
    rabbit-mq:
      enabled: true
      uri: "amqp://user:password@localhost/"`)
	return configurationsCoreBytes, nil
}

func mockReadfilebytesFormatError(filePath string) ([]byte, error) {
	configurationsCoreBytes := []byte(`{{}`)
	return configurationsCoreBytes, nil
}

func mockReadfilebytesError(filePath string) ([]byte, error) {
	return nil, errors.New("mockErrorReadFileBytes")
}

func mockReadDirError(dirPath string) ([]string, error) {
	return nil, errors.New("mockErrorReadFileBytes")
}

func mockReadDirValid(dirPath string) ([]string, error) {
	return []string{""}, nil
}

func mockReadfilebytesBeelzebubServiceConfiguration(filePath string) ([]byte, error) {
	beelzebubServiceConfiguration := []byte(`
apiVersion: "v1"
protocol: "http"
address: ":8080"
commands:
  - regex: "wp-admin"
    handler: "login"
    headers:
      - "Content-Type: text/html"`)
	return beelzebubServiceConfiguration, nil
}

func TestReadConfigurationsCoreError(t *testing.T) {
	configurationsParser := Init("mockConfigurationsCorePath", "mockConfigurationsServicesDirectory")

	configurationsParser.readFileBytesByFilePathDependency = mockReadfilebytesError
	beelzebubCoreConfigurations, err := configurationsParser.ReadConfigurationsCore()

	assert.Nil(t, beelzebubCoreConfigurations)
	assert.Error(t, err)
	assert.Equal(t, "in file mockConfigurationsCorePath: mockErrorReadFileBytes", err.Error())

	configurationsParser.readFileBytesByFilePathDependency = mockReadfilebytesFormatError

	beelzebubCoreConfigurations, err = configurationsParser.ReadConfigurationsCore()
	assert.Nil(t, beelzebubCoreConfigurations)
	assert.Error(t, err)
	assert.Equal(t, "in file mockConfigurationsCorePath: yaml: line 1: did not find expected ',' or '}'", err.Error())
}

func TestReadConfigurationsCoreValid(t *testing.T) {
	configurationsParser := Init("", "")
	configurationsParser.readFileBytesByFilePathDependency = mockReadfilebytesConfigurationsCore

	coreConfigurations, err := configurationsParser.ReadConfigurationsCore()
	assert.Nil(t, err)
	assert.NotNil(t, coreConfigurations.Core)
	assert.NotNil(t, coreConfigurations.Core.Logging)
	assert.Equal(t, coreConfigurations.Core.Logging.Debug, false)
	assert.Equal(t, coreConfigurations.Core.Logging.LogDisableTimestamp, true)
	assert.Equal(t, coreConfigurations.Core.Logging.DebugReportCaller, false)
	assert.Equal(t, coreConfigurations.Core.Logging.LogsPath, "./logs")
	assert.Equal(t, coreConfigurations.Core.Tracings.RabbitMQ.Enabled, true)
	assert.Equal(t, coreConfigurations.Core.Tracings.RabbitMQ.URI, "amqp://user:password@localhost/")
}

func TestReadConfigurationsServicesFail(t *testing.T) {
	configurationsParser := Init("", "")

	configurationsParser.readFileBytesByFilePathDependency = mockReadfilebytesError
	configurationsParser.gelAllFilesNameByDirNameDependency = mockReadDirError

	beelzebubServiceConfiguration, err := configurationsParser.ReadConfigurationsServices()
	assert.Nil(t, beelzebubServiceConfiguration)
	assert.Error(t, err)
}

func TestReadConfigurationsServicesValid(t *testing.T) {
	configurationsParser := Init("", "")

	configurationsParser.readFileBytesByFilePathDependency = mockReadfilebytesBeelzebubServiceConfiguration
	configurationsParser.gelAllFilesNameByDirNameDependency = mockReadDirValid

	beelzebubServicesConfiguration, err := configurationsParser.ReadConfigurationsServices()

	firstBeelzebubServiceConfiguration := beelzebubServicesConfiguration[0]

	assert.Nil(t, err)
	assert.Equal(t, firstBeelzebubServiceConfiguration.Protocol, "http")
	assert.Equal(t, firstBeelzebubServiceConfiguration.ApiVersion, "v1")
	assert.Equal(t, firstBeelzebubServiceConfiguration.Address, ":8080")
	assert.Equal(t, len(firstBeelzebubServiceConfiguration.Commands), 1)
	assert.Equal(t, len(firstBeelzebubServiceConfiguration.Commands), 1)
	assert.Equal(t, firstBeelzebubServiceConfiguration.Commands[0].Regex, "wp-admin")
	assert.Equal(t, firstBeelzebubServiceConfiguration.Commands[0].Handler, "login")
	assert.Equal(t, len(firstBeelzebubServiceConfiguration.Commands[0].Headers), 1)
	assert.Equal(t, firstBeelzebubServiceConfiguration.Commands[0].Headers[0], "Content-Type: text/html")
}
