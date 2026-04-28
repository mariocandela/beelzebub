package cli

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatBool(t *testing.T) {
	if got := formatBool(true); got != "enabled" {
		t.Errorf("formatBool(true) = %q, want %q", got, "enabled")
	}
	if got := formatBool(false); got != "disabled" {
		t.Errorf("formatBool(false) = %q, want %q", got, "disabled")
	}
}

func TestFormatOptional(t *testing.T) {
	if got := formatOptional(""); got != "(not set)" {
		t.Errorf("formatOptional(\"\") = %q, want %q", got, "(not set)")
	}
	if got := formatOptional("value"); got != "value" {
		t.Errorf("formatOptional(\"value\") = %q, want %q", got, "value")
	}
}

func TestPrintSectionAndField(t *testing.T) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printSection("Test Title", "Test Detail")
	printField("Test Name", "Test Value")

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	out := buf.String()

	if !strings.Contains(out, "Test Title: Test Detail\n") {
		t.Errorf("printSection output incorrect, got: %q", out)
	}
	if !strings.Contains(out, "  Test Name:         Test Value\n") {
		t.Errorf("printField output incorrect, got: %q", out)
	}
}

func TestValidateConfigurations_Success(t *testing.T) {
	rootConfCore = "../configurations/beelzebub.yaml"
	rootConfServices = "../configurations/services/"

	oldStdout := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	err := validateConfigurations(nil, nil)

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestValidateConfigurations_InvalidCoreYaml(t *testing.T) {
	tmpDir := t.TempDir()
	corePath := filepath.Join(tmpDir, "core.yaml")
	os.WriteFile(corePath, []byte("invalid: yaml: :"), 0644)

	rootConfCore = corePath
	rootConfServices = "../configurations/services/"

	err := validateConfigurations(nil, nil)
	if err == nil {
		t.Error("expected error for invalid core config yaml")
	}
}

func TestValidateConfigurations_InvalidServicesYaml(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "svc.yaml"), []byte("invalid: yaml: :"), 0644)

	rootConfCore = "../configurations/beelzebub.yaml"
	rootConfServices = tmpDir

	err := validateConfigurations(nil, nil)
	if err == nil {
		t.Error("expected error for invalid services config yaml")
	}
}

func TestValidateConfigurations_UnknownProtocol(t *testing.T) {
	tmpDir := t.TempDir()
	yamlContent := `
apiVersion: v1
protocol: unknown_protocol
address: ":8080"
`
	os.WriteFile(filepath.Join(tmpDir, "svc.yaml"), []byte(yamlContent), 0644)

	rootConfCore = "../configurations/beelzebub.yaml"
	rootConfServices = tmpDir

	err := validateConfigurations(nil, nil)
	if err == nil {
		t.Error("expected error for unknown protocol")
	} else if !strings.Contains(err.Error(), "unknown protocol") {
		t.Errorf("expected error to mention unknown protocol, got: %v", err)
	}
}

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

	rootConfCore = tmpDir + "/beelzebub.yaml"
	rootConfServices = servicesDir

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

	rootConfCore = tmpDir + "/beelzebub.yaml"
	rootConfServices = servicesDir

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

	rootConfCore = tmpDir + "/beelzebub.yaml"
	rootConfServices = servicesDir

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

	rootConfCore = "/nonexistent/path/beelzebub.yaml"
	rootConfServices = servicesDir

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

	rootConfCore = tmpDir + "/beelzebub.yaml"
	rootConfServices = servicesDir

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

	rootConfCore = tmpDir + "/beelzebub.yaml"
	rootConfServices = servicesDir

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

func TestValidateConfigurations_ServicesPathIsFile(t *testing.T) {
	tmpDir := t.TempDir()

	coreYAML := `
core:
  logging:
    debug: false
`
	assert.NoError(t, os.WriteFile(tmpDir+"/beelzebub.yaml", []byte(coreYAML), 0644))
	assert.NoError(t, os.WriteFile(tmpDir+"/services.yaml", []byte("not a directory"), 0644))

	err := runValidate(tmpDir+"/beelzebub.yaml", tmpDir+"/services.yaml")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "services config")
}
