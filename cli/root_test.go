package cli

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestPrintVersion(t *testing.T) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Set test values
	Version = "1.0.0"
	CommitSHA = "abcdef"
	BuildDate = "2024-01-01"

	printVersion(versionCmd, nil)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	out := buf.String()

	if !strings.Contains(out, "beelzebub 1.0.0") {
		t.Errorf("expected version output to contain 'beelzebub 1.0.0', got: %s", out)
	}
	if !strings.Contains(out, "commit:     abcdef") {
		t.Errorf("expected version output to contain commit hash, got: %s", out)
	}
	if !strings.Contains(out, "build date: 2024-01-01") {
		t.Errorf("expected version output to contain build date, got: %s", out)
	}
}

func TestPrintVersion_DevFallback(t *testing.T) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Set test values to trigger dev fallback
	Version = "dev"
	CommitSHA = "unknown"
	BuildDate = "unknown"

	printVersion(versionCmd, nil)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	out := buf.String()

	if !strings.Contains(out, "beelzebub") {
		t.Errorf("expected version output to contain 'beelzebub', got: %s", out)
	}
	if !strings.Contains(out, "commit:") {
		t.Errorf("expected version output to contain commit, got: %s", out)
	}
}

func TestListPlugins(t *testing.T) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	listPlugins(pluginListCmd, nil)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	out := buf.String()

	if !strings.Contains(out, "No plugins registered.") && !strings.Contains(out, "NAME") {
		t.Errorf("unexpected plugin list output: %s", out)
	}
}

func TestRootCmd_PersistentPreRunE(t *testing.T) {
	// Test valid log level
	rootLogLevel = "debug"
	err := rootCmd.PersistentPreRunE(rootCmd, nil)
	if err != nil {
		t.Errorf("expected no error for valid log level, got: %v", err)
	}

	// Test invalid log level
	rootLogLevel = "invalid-level"
	err = rootCmd.PersistentPreRunE(rootCmd, nil)
	if err == nil {
		t.Errorf("expected error for invalid log level, got nil")
	}
}

func TestExecute(t *testing.T) {
	// Cannot easily test full Execute but we can just call the command manually with help
	os.Args = []string{"beelzebub", "--help"}
	err := Execute()
	if err != nil {
		t.Errorf("expected no error from Execute with help, got %v", err)
	}
}
