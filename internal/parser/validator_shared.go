package parser

import (
	"fmt"
	"os"
	"regexp"
)

// ValidateTLSConfig checks that TLS cert and key paths are both set or both empty,
// and warns if the referenced files do not exist
func ValidateTLSConfig(certPath, keyPath string) []ValidationIssue {
	var issues []ValidationIssue

	if (certPath != "" && keyPath == "") || (certPath == "" && keyPath != "") {
		issues = append(issues, ValidationIssue{
			Level:   LevelError,
			Message: "both tlsCertPath and tlsKeyPath must be set for TLS, or neither",
		})
	}

	if certPath != "" && keyPath != "" {
		if _, err := os.Stat(certPath); os.IsNotExist(err) {
			issues = append(issues, ValidationIssue{
				Level:   LevelWarning,
				Message: "tlsCertPath file does not exist",
			})
		}
		if _, err := os.Stat(keyPath); os.IsNotExist(err) {
			issues = append(issues, ValidationIssue{
				Level:   LevelWarning,
				Message: "tlsKeyPath file does not exist",
			})
		}
	}

	return issues
}

// ValidatePasswordRegex checks that the password regex is non-empty and valid
func ValidatePasswordRegex(passwordRegex, protocolName string) []ValidationIssue {
	var issues []ValidationIssue

	if passwordRegex == "" {
		issues = append(issues, ValidationIssue{
			Level:   LevelError,
			Message: fmt.Sprintf("passwordRegex is required for %s protocol", protocolName),
		})
	} else if _, err := regexp.Compile(passwordRegex); err != nil {
		issues = append(issues, ValidationIssue{
			Level:   LevelError,
			Message: fmt.Sprintf("passwordRegex is not a valid regex: %v", err),
		})
	}

	return issues
}
