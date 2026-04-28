package cli

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateConfigurations_ValidConfigs(t *testing.T) {
	tmpDir := t.TempDir()
	servicesDir := t.TempDir()

	coreYAML := `
core:
  logging:
    debug: false
`
	err := os.WriteFile(tmpDir+"/beelzebub.yaml", []byte(coreYAML), 0644)
	assert.NoError(t, err)

	svcYAML := `
apiVersion: v1
protocol: http
address: ":8080"
deadlineTimeoutSeconds: 60
commands:
  - regex: "^GET /"
    handler: "handler"
fallbackCommand:
  handler: "fallback"
`
	err = os.WriteFile(servicesDir+"/http-8080.yaml", []byte(svcYAML), 0644)
	assert.NoError(t, err)

	validateConfCore = tmpDir + "/beelzebub.yaml"
	validateConfServices = servicesDir

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = validateConfigurations(nil, nil)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)

	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "0 errors, 0 warnings")
}

func TestValidateConfigurations_InvalidServiceConfig(t *testing.T) {
	tmpDir := t.TempDir()
	servicesDir := t.TempDir()

	coreYAML := `
core:
  logging:
    debug: false
`
	err := os.WriteFile(tmpDir+"/beelzebub.yaml", []byte(coreYAML), 0644)
	assert.NoError(t, err)

	svcYAML := `
apiVersion: v1
protocol: ftp
address: ":8080"
`
	err = os.WriteFile(servicesDir+"/bad.yaml", []byte(svcYAML), 0644)
	assert.NoError(t, err)

	validateConfCore = tmpDir + "/beelzebub.yaml"
	validateConfServices = servicesDir

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = validateConfigurations(nil, nil)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validation failed with 1 error(s)")
	assert.Contains(t, buf.String(), "invalid protocol")
}

func TestValidateConfigurations_CoreConfigParseError(t *testing.T) {
	servicesDir := t.TempDir()
	tmpDir := t.TempDir()

	malformedYAML := `
core:
  logging:
    debug: [this is not valid yaml
`
	err := os.WriteFile(tmpDir+"/beelzebub.yaml", []byte(malformedYAML), 0644)
	assert.NoError(t, err)

	validateConfCore = tmpDir + "/beelzebub.yaml"
	validateConfServices = servicesDir

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = validateConfigurations(nil, nil)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validation failed")
	assert.Contains(t, buf.String(), "failed to read core config")
}

func TestValidateConfigurations_MissingCoreConfig(t *testing.T) {
	servicesDir := t.TempDir()

	validateConfCore = "/nonexistent/path/beelzebub.yaml"
	validateConfServices = servicesDir

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := validateConfigurations(nil, nil)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)

	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "0 errors, 0 warnings")
}

func TestValidateConfigurations_MalformedServiceConfig(t *testing.T) {
	tmpDir := t.TempDir()
	servicesDir := t.TempDir()

	coreYAML := `
core:
  logging:
    debug: false
`
	err := os.WriteFile(tmpDir+"/beelzebub.yaml", []byte(coreYAML), 0644)
	assert.NoError(t, err)

	badSvcYAML := `
apiVersion: v1
protocol: ssh
address: ":22"
  this is broken indentation
`
	err = os.WriteFile(servicesDir+"/broken.yaml", []byte(badSvcYAML), 0644)
	assert.NoError(t, err)

	validateConfCore = tmpDir + "/beelzebub.yaml"
	validateConfServices = servicesDir

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = validateConfigurations(nil, nil)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)

	assert.Error(t, err)
	output := buf.String()
	assert.True(t, strings.Contains(output, "FAIL broken.yaml") || strings.Contains(output, "YAML"))
}

func TestValidateConfigurations_EmptyServicesDir(t *testing.T) {
	tmpDir := t.TempDir()
	servicesDir := t.TempDir()

	coreYAML := `
core:
  logging:
    debug: false
`
	err := os.WriteFile(tmpDir+"/beelzebub.yaml", []byte(coreYAML), 0644)
	assert.NoError(t, err)

	validateConfCore = tmpDir + "/beelzebub.yaml"
	validateConfServices = servicesDir

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = validateConfigurations(nil, nil)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)

	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "0 errors, 0 warnings")
}
