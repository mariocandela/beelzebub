package parser

import (
	"errors"
	"os"
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
      uri: "amqp://user:password@localhost/"
  beelzebub-cloud:
    enabled: true
    uri: "amqp://user:password@localhost/"
    auth-token: "iejfdjsl-aosdajosoidaj-dunfkjnfkjsdnkn"`)
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
tlsCertPath: "/tmp/cert.crt"
tlsKeyPath: "/tmp/cert.key"
commands:
  - regex: "wp-admin"
    handler: "login"
    headers:
      - "Content-Type: text/html"
  - name: "wp-admin"
    regex: "wp-admin"
    handler: "login"
    headers:
      - "Content-Type: text/html"
fallbackCommand:
  handler: "404 Not Found!"
  statusCode: 404
plugin:
  openAISecretKey: "qwerty"
  llmModel: "llama3"
  llmProvider: "ollama"
  host: "localhost:1563"
  prompt: "hello world"
`)
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
	assert.Equal(t, coreConfigurations.Core.BeelzebubCloud.Enabled, true)
	assert.Equal(t, coreConfigurations.Core.BeelzebubCloud.URI, "amqp://user:password@localhost/")
	assert.Equal(t, coreConfigurations.Core.BeelzebubCloud.AuthToken, "iejfdjsl-aosdajosoidaj-dunfkjnfkjsdnkn")
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
	assert.Nil(t, err)

	firstBeelzebubServiceConfiguration := beelzebubServicesConfiguration[0]

	assert.Equal(t, firstBeelzebubServiceConfiguration.Protocol, "http")
	assert.Equal(t, firstBeelzebubServiceConfiguration.ApiVersion, "v1")
	assert.Equal(t, firstBeelzebubServiceConfiguration.Address, ":8080")
	assert.Equal(t, len(firstBeelzebubServiceConfiguration.Commands), 2)
	assert.Equal(t, len(firstBeelzebubServiceConfiguration.Commands), 2)
	assert.Equal(t, firstBeelzebubServiceConfiguration.Commands[0].Regex, "wp-admin")
	assert.Equal(t, firstBeelzebubServiceConfiguration.Commands[0].Handler, "login")
	assert.Equal(t, len(firstBeelzebubServiceConfiguration.Commands[0].Headers), 1)
	assert.Equal(t, firstBeelzebubServiceConfiguration.Commands[0].Headers[0], "Content-Type: text/html")
	assert.Equal(t, firstBeelzebubServiceConfiguration.Commands[1].Name, "wp-admin")
	assert.Equal(t, firstBeelzebubServiceConfiguration.FallbackCommand.Handler, "404 Not Found!")
	assert.Equal(t, firstBeelzebubServiceConfiguration.FallbackCommand.StatusCode, 404)
	assert.Equal(t, firstBeelzebubServiceConfiguration.Plugin.OpenAISecretKey, "qwerty")
	assert.Equal(t, firstBeelzebubServiceConfiguration.Plugin.LLMModel, "llama3")
	assert.Equal(t, firstBeelzebubServiceConfiguration.Plugin.LLMProvider, "ollama")
	assert.Equal(t, firstBeelzebubServiceConfiguration.Plugin.Host, "localhost:1563")
	assert.Equal(t, firstBeelzebubServiceConfiguration.Plugin.Prompt, "hello world")
	assert.Equal(t, firstBeelzebubServiceConfiguration.TLSCertPath, "/tmp/cert.crt")
	assert.Equal(t, firstBeelzebubServiceConfiguration.TLSKeyPath, "/tmp/cert.key")
}

func TestGelAllFilesNameByDirName(t *testing.T) {

	var dir = t.TempDir()

	files, err := gelAllFilesNameByDirName(dir)

	assert.Nil(t, err)
	assert.Equal(t, 0, len(files))
}

func TestGelAllFilesNameByDirNameFiles(t *testing.T) {

	var dir = t.TempDir()

	testFiles := []string{"file1.yaml", "file2.yaml", "file3.txt", "subdir", "file4.yaml"}
	for _, filename := range testFiles {
		filePath := dir + "/" + filename
		file, err := os.Create(filePath)
		assert.NoError(t, err)
		file.Close()
	}

	files, err := gelAllFilesNameByDirName(dir)

	assert.Nil(t, err)
	assert.Equal(t, 3, len(files))
}

func TestGelAllFilesNameByDirNameError(t *testing.T) {

	files, err := gelAllFilesNameByDirName("nosuchfile")

	assert.Nil(t, files)
	// Windows and Linux return slightly different error strings, but share a common prefix, so check for that.
	assert.Contains(t, err.Error(), "open nosuchfile: ")
}

func TestReadFileBytesByFilePath(t *testing.T) {

	var dir = t.TempDir()
	filePath := dir + "/test.yaml"

	f, err := os.Create(filePath)
	assert.NoError(t, err)
	f.Close()

	bytes, err := readFileBytesByFilePath(filePath)
	assert.NoError(t, err)

	assert.Equal(t, "", string(bytes))
}
