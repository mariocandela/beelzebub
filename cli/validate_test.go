package cli

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
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
