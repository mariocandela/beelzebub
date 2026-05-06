package parser

import (
	"fmt"
	"os"
	"regexp"
)

// ValidateTLSConfig checks that TLS cert and key paths are both set or both empty,
// and warns if the referenced files do not exist
func ValidateTLSConfig(certPath, keyPath, filename string) []ValidationIssue {
	var issues []ValidationIssue

	if (certPath != "" && keyPath == "") || (certPath == "" && keyPath != "") {
		issues = append(issues, ValidationIssue{
			Level:    LevelError,
			Message:  "both tlsCertPath and tlsKeyPath must be set for TLS, or neither",
			Filename: filename,
		})
	}

	if certPath != "" && keyPath != "" {
		if _, err := os.Stat(certPath); os.IsNotExist(err) {
			issues = append(issues, ValidationIssue{
				Level:    LevelWarning,
				Message:  "tlsCertPath file does not exist",
				Filename: filename,
			})
		}
		if _, err := os.Stat(keyPath); os.IsNotExist(err) {
			issues = append(issues, ValidationIssue{
				Level:    LevelWarning,
				Message:  "tlsKeyPath file does not exist",
				Filename: filename,
			})
		}
	}

	return issues
}

func ValidatePasswordRegex(passwordRegex, protocolName, filename string) []ValidationIssue {
	var issues []ValidationIssue

	if passwordRegex == "" {
		issues = append(issues, ValidationIssue{
			Level:    LevelError,
			Message:  fmt.Sprintf("passwordRegex is required for %s protocol", protocolName),
			Filename: filename,
		})
	} else if _, err := regexp.Compile(passwordRegex); err != nil {
		issues = append(issues, ValidationIssue{
			Level:    LevelError,
			Message:  fmt.Sprintf("passwordRegex is not a valid regex: %v", err),
			Filename: filename,
		})
	}

	return issues
}
