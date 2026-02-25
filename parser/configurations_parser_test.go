package parser

import (
	"errors"
	"os"
	"regexp"
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
tools:
  - name: "tool:user-account-manager"
    description: "Tool for querying and modifying user account details. Requires administrator privileges."
    params:
      - name: "user_id"
        description: "The ID of the user account to manage."
      - name: "action"
        description: "The action to perform on the user account, possible values are: get_details, reset_password, deactivate_account"
    handler: "reset_password ok"
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
  inputValidationEnabled: true
  inputValidationPrompt: "hello world"
  outputValidationEnabled: true
  outputValidationPrompt: "hello world"
`)
	return beelzebubServiceConfiguration, nil
}

func mockReadfilebytesBeelzebubServiceConfigurationDefaultValues(filePath string) ([]byte, error) {
	beelzebubServiceConfiguration := []byte(``)
	return beelzebubServiceConfiguration, nil
}

func mockReadfilebytesToolWithReadOnlyAnnotation(filePath string) ([]byte, error) {
	beelzebubServiceConfiguration := []byte(`
apiVersion: "v1"
protocol: "mcp"
address: ":8000"
tools:
  - name: "tool:query-logs"
    description: "Query system logs for analysis"
    annotations:
      title: "Query Logs"
      readOnlyHint: true
    params:
      - name: "filter"
        description: "Log filter criteria"
    handler: "log_query_handler"
`)
	return beelzebubServiceConfiguration, nil
}

func mockReadfilebytesToolWithDestructiveAnnotation(filePath string) ([]byte, error) {
	beelzebubServiceConfiguration := []byte(`
apiVersion: "v1"
protocol: "mcp"
address: ":8000"
tools:
  - name: "tool:delete-user"
    description: "Delete a user account permanently"
    annotations:
      title: "Delete User"
      destructiveHint: true
    params:
      - name: "user_id"
        description: "The user ID to delete"
    handler: "delete_user_handler"
`)
	return beelzebubServiceConfiguration, nil
}

func mockReadfilebytesToolWithMultipleAnnotations(filePath string) ([]byte, error) {
	beelzebubServiceConfiguration := []byte(`
apiVersion: "v1"
protocol: "mcp"
address: ":8000"
tools:
  - name: "tool:update-config"
    description: "Update system configuration"
    annotations:
      title: "Update Config"
      destructiveHint: true
      idempotentHint: true
      openWorldHint: false
    params:
      - name: "config_key"
        description: "Configuration key to update"
    handler: "update_config_handler"
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
	assert.Equal(t, firstBeelzebubServiceConfiguration.Commands[0].RegexStr, "wp-admin")
	assert.Equal(t, firstBeelzebubServiceConfiguration.Commands[0].Regex.String(), "wp-admin")
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
	assert.Equal(t, firstBeelzebubServiceConfiguration.TLSKeyPath, "/tmp/cert.key")
	assert.Equal(t, len(firstBeelzebubServiceConfiguration.Tools), 1)
	assert.Equal(t, firstBeelzebubServiceConfiguration.Tools[0].Name, "tool:user-account-manager")
	assert.Equal(t, firstBeelzebubServiceConfiguration.Tools[0].Description, "Tool for querying and modifying user account details. Requires administrator privileges.")
	assert.Equal(t, len(firstBeelzebubServiceConfiguration.Tools[0].Params), 2)
	assert.Equal(t, firstBeelzebubServiceConfiguration.Tools[0].Params[0].Name, "user_id")
	assert.Equal(t, firstBeelzebubServiceConfiguration.Tools[0].Params[0].Description, "The ID of the user account to manage.")
	assert.Equal(t, firstBeelzebubServiceConfiguration.Tools[0].Handler, "reset_password ok")
}

func TestReadConfigurationsServicesGenerateHashCode(t *testing.T) {
	configurationsParser := Init("", "")

	configurationsParser.readFileBytesByFilePathDependency = mockReadfilebytesBeelzebubServiceConfiguration
	configurationsParser.gelAllFilesNameByDirNameDependency = mockReadDirValid

	beelzebubServicesConfiguration, err := configurationsParser.ReadConfigurationsServices()

	hashCode, errHashCode := beelzebubServicesConfiguration[0].HashCode()

	assert.Nil(t, err)
	assert.Nil(t, errHashCode)
	assert.Equal(t, hashCode, "9c349217fdf25f8a1751c33de9e06799a6c96fa996c2dba40df6d2c34c3025a0")
}

func TestReadConfigurationsPluginGuardrailsValid(t *testing.T) {
	configurationsParser := Init("", "")

	configurationsParser.readFileBytesByFilePathDependency = mockReadfilebytesBeelzebubServiceConfiguration
	configurationsParser.gelAllFilesNameByDirNameDependency = mockReadDirValid

	beelzebubServicesConfiguration, err := configurationsParser.ReadConfigurationsServices()
	assert.Nil(t, err)

	firstBeelzebubServiceConfiguration := beelzebubServicesConfiguration[0]

	assert.Equal(t, firstBeelzebubServiceConfiguration.Plugin.InputValidationEnabled, true)
	assert.Equal(t, firstBeelzebubServiceConfiguration.Plugin.InputValidationPrompt, "hello world")
	assert.Equal(t, firstBeelzebubServiceConfiguration.Plugin.OutputValidationEnabled, true)
	assert.Equal(t, firstBeelzebubServiceConfiguration.Plugin.OutputValidationPrompt, "hello world")
}

func TestReadConfigurationsDefaultValues(t *testing.T) {

	configurationsParser := Init("", "")

	configurationsParser.readFileBytesByFilePathDependency = mockReadfilebytesBeelzebubServiceConfigurationDefaultValues
	configurationsParser.gelAllFilesNameByDirNameDependency = mockReadDirValid
	beelzebubServicesConfiguration, err := configurationsParser.ReadConfigurationsServices()
	assert.Nil(t, err)

	firstBeelzebubServiceConfiguration := beelzebubServicesConfiguration[0]
	assert.Equal(t, firstBeelzebubServiceConfiguration.Protocol, "")
	assert.Equal(t, firstBeelzebubServiceConfiguration.ApiVersion, "")
	assert.Equal(t, firstBeelzebubServiceConfiguration.Address, "")
	assert.Equal(t, len(firstBeelzebubServiceConfiguration.Commands), 0)
	assert.Equal(t, firstBeelzebubServiceConfiguration.FallbackCommand.Handler, "")
	assert.Equal(t, firstBeelzebubServiceConfiguration.FallbackCommand.StatusCode, 0)
	assert.Equal(t, firstBeelzebubServiceConfiguration.Plugin.OpenAISecretKey, "")
	assert.Equal(t, firstBeelzebubServiceConfiguration.Plugin.LLMModel, "")
	assert.Equal(t, firstBeelzebubServiceConfiguration.Plugin.LLMProvider, "")
	assert.Equal(t, firstBeelzebubServiceConfiguration.Plugin.Host, "")
	assert.Equal(t, firstBeelzebubServiceConfiguration.Plugin.Prompt, "")
	assert.Equal(t, firstBeelzebubServiceConfiguration.TLSCertPath, "")
	assert.Equal(t, firstBeelzebubServiceConfiguration.TLSKeyPath, "")
	assert.Equal(t, len(firstBeelzebubServiceConfiguration.Tools), 0)
	assert.Equal(t, firstBeelzebubServiceConfiguration.Plugin.InputValidationEnabled, false)
	assert.Equal(t, firstBeelzebubServiceConfiguration.Plugin.InputValidationPrompt, "")
	assert.Equal(t, firstBeelzebubServiceConfiguration.Plugin.OutputValidationEnabled, false)
	assert.Equal(t, firstBeelzebubServiceConfiguration.Plugin.OutputValidationPrompt, "")
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

func TestCompileCommandRegex(t *testing.T) {
	tests := []struct {
		name          string
		config        BeelzebubServiceConfiguration
		expectedError bool
	}{
		{
			name: "Valid Regex",
			config: BeelzebubServiceConfiguration{
				Commands: []Command{
					{RegexStr: "^/api/v1/.*$"},
					{RegexStr: "wp-admin"},
				},
			},
			expectedError: false,
		},
		{
			name: "Empty Regex",
			config: BeelzebubServiceConfiguration{
				Commands: []Command{
					{RegexStr: ""},
					{RegexStr: ""},
				},
			},
			expectedError: false,
		},
		{
			name: "Invalid Regex",
			config: BeelzebubServiceConfiguration{
				Commands: []Command{
					{RegexStr: "["},
				},
			},
			expectedError: true,
		},
		{
			name: "Mixed valid and Invalid Regex",
			config: BeelzebubServiceConfiguration{
				Commands: []Command{
					{RegexStr: "^/api/v1/.*$"},
					{RegexStr: "["},
					{RegexStr: "test"},
				},
			},
			expectedError: true,
		},
		{
			name:          "No commands",
			config:        BeelzebubServiceConfiguration{},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.CompileCommandRegex()

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				for _, command := range tt.config.Commands {
					if command.RegexStr != "" {
						assert.NotNil(t, command.Regex)
						_, err := regexp.Compile(command.RegexStr)
						assert.NoError(t, err)

					} else {
						assert.Nil(t, command.Regex)
					}
				}
			}
		})
	}
}

func TestToolAnnotationsReadOnlyHint(t *testing.T) {
	configurationsParser := Init("", "")
	configurationsParser.readFileBytesByFilePathDependency = mockReadfilebytesToolWithReadOnlyAnnotation
	configurationsParser.gelAllFilesNameByDirNameDependency = mockReadDirValid

	beelzebubServicesConfiguration, err := configurationsParser.ReadConfigurationsServices()
	assert.Nil(t, err)

	tool := beelzebubServicesConfiguration[0].Tools[0]
	assert.Equal(t, "tool:query-logs", tool.Name)
	assert.Equal(t, "Query system logs for analysis", tool.Description)

	// Verify annotations are parsed correctly
	assert.NotNil(t, tool.Annotations)
	assert.Equal(t, "Query Logs", tool.Annotations.Title)
	assert.NotNil(t, tool.Annotations.ReadOnlyHint)
	assert.True(t, *tool.Annotations.ReadOnlyHint)

	// Other hints should be nil (not specified in YAML)
	assert.Nil(t, tool.Annotations.DestructiveHint)
	assert.Nil(t, tool.Annotations.IdempotentHint)
	assert.Nil(t, tool.Annotations.OpenWorldHint)
}

func TestToolAnnotationsDestructiveHint(t *testing.T) {
	configurationsParser := Init("", "")
	configurationsParser.readFileBytesByFilePathDependency = mockReadfilebytesToolWithDestructiveAnnotation
	configurationsParser.gelAllFilesNameByDirNameDependency = mockReadDirValid

	beelzebubServicesConfiguration, err := configurationsParser.ReadConfigurationsServices()
	assert.Nil(t, err)

	tool := beelzebubServicesConfiguration[0].Tools[0]
	assert.Equal(t, "tool:delete-user", tool.Name)

	// Verify annotations are parsed correctly
	assert.NotNil(t, tool.Annotations)
	assert.Equal(t, "Delete User", tool.Annotations.Title)
	assert.NotNil(t, tool.Annotations.DestructiveHint)
	assert.True(t, *tool.Annotations.DestructiveHint)

	// ReadOnlyHint should be nil (not specified)
	assert.Nil(t, tool.Annotations.ReadOnlyHint)
}

func TestToolAnnotationsMultipleHints(t *testing.T) {
	configurationsParser := Init("", "")
	configurationsParser.readFileBytesByFilePathDependency = mockReadfilebytesToolWithMultipleAnnotations
	configurationsParser.gelAllFilesNameByDirNameDependency = mockReadDirValid

	beelzebubServicesConfiguration, err := configurationsParser.ReadConfigurationsServices()
	assert.Nil(t, err)

	tool := beelzebubServicesConfiguration[0].Tools[0]
	assert.Equal(t, "tool:update-config", tool.Name)

	// Verify all annotations are parsed correctly
	assert.NotNil(t, tool.Annotations)
	assert.Equal(t, "Update Config", tool.Annotations.Title)

	assert.NotNil(t, tool.Annotations.DestructiveHint)
	assert.True(t, *tool.Annotations.DestructiveHint)

	assert.NotNil(t, tool.Annotations.IdempotentHint)
	assert.True(t, *tool.Annotations.IdempotentHint)

	assert.NotNil(t, tool.Annotations.OpenWorldHint)
	assert.False(t, *tool.Annotations.OpenWorldHint)

	// ReadOnlyHint should be nil (not specified)
	assert.Nil(t, tool.Annotations.ReadOnlyHint)
}

func TestToolWithoutAnnotations(t *testing.T) {
	configurationsParser := Init("", "")
	// Uses existing mock that has tools without annotations
	configurationsParser.readFileBytesByFilePathDependency = mockReadfilebytesBeelzebubServiceConfiguration
	configurationsParser.gelAllFilesNameByDirNameDependency = mockReadDirValid

	beelzebubServicesConfiguration, err := configurationsParser.ReadConfigurationsServices()
	assert.Nil(t, err)

	tool := beelzebubServicesConfiguration[0].Tools[0]
	assert.Equal(t, "tool:user-account-manager", tool.Name)

	// Annotations should be nil for backward compatibility
	assert.Nil(t, tool.Annotations)

	// Tool should still be fully functional
	assert.Equal(t, "Tool for querying and modifying user account details. Requires administrator privileges.", tool.Description)
	assert.Equal(t, 2, len(tool.Params))
	assert.Equal(t, "reset_password ok", tool.Handler)
}

func TestToolAnnotationsHashCodeStability(t *testing.T) {
	configurationsParser := Init("", "")
	// Use existing mock without annotations
	configurationsParser.readFileBytesByFilePathDependency = mockReadfilebytesBeelzebubServiceConfiguration
	configurationsParser.gelAllFilesNameByDirNameDependency = mockReadDirValid

	beelzebubServicesConfiguration, err := configurationsParser.ReadConfigurationsServices()
	assert.Nil(t, err)

	// Existing hash from TestReadConfigurationsServicesGenerateHashCode
	// This ensures that adding the Annotations field with omitempty doesn't change
	// the hash for configs that don't use annotations
	hashCode, errHashCode := beelzebubServicesConfiguration[0].HashCode()
	assert.Nil(t, errHashCode)
	assert.Equal(t, "9c349217fdf25f8a1751c33de9e06799a6c96fa996c2dba40df6d2c34c3025a0", hashCode)
}
