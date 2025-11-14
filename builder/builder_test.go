package builder

import (
	"os"
	"testing"

	"github.com/mariocandela/beelzebub/v3/parser"
	"github.com/stretchr/testify/assert"
)

func TestBuilderClose_LogFile(t *testing.T) {
	// Create a temporary directory for the test
	tmpDir := t.TempDir()
	logFilePath := tmpDir + "/test.log"

	// Create a builder instance
	builder := NewBuilder()

	// Build logger which opens a log file
	loggingConfig := parser.Logging{
		Debug:               false,
		DebugReportCaller:   false,
		LogDisableTimestamp: true,
		LogsPath:            logFilePath,
	}

	err := builder.buildLogger(loggingConfig)
	assert.NoError(t, err)
	assert.NotNil(t, builder.logsFile)

	// Verify the log file exists and is open
	fileInfo, err := os.Stat(logFilePath)
	assert.NoError(t, err)
	assert.NotNil(t, fileInfo)

	// Close the builder
	err = builder.Close()
	assert.NoError(t, err)

	// Verify the log file is closed by attempting to write to it
	// Writing to a closed file should return an error
	_, err = builder.logsFile.WriteString("test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "file already closed")
}

func TestBuilderClose_NoLogFile(t *testing.T) {
	// Create a builder without opening a log file
	builder := NewBuilder()

	// Close should succeed even without a log file
	err := builder.Close()
	assert.NoError(t, err)
}

func TestBuilderClose_NilLogFile(t *testing.T) {
	// Create a builder with explicitly nil log file
	builder := &Builder{
		logsFile: nil,
	}

	// Close should succeed with nil log file
	err := builder.Close()
	assert.NoError(t, err)
}
