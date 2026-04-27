package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"slices"
	"strconv"
	"strings"
)

const (
	LevelError   = "error"
	LevelWarning = "warning"
)

type ValidationIssue struct {
	Level    string
	Message  string
	Filename string
}

type ValidationResult struct {
	Filename string
	Issues   []ValidationIssue
}

type ValidateResult struct {
	Results       []ValidationResult
	TotalErrors   int
	TotalWarnings int
}

type ServiceValidator interface {
	Name() string
	Validate(config BeelzebubServiceConfiguration) []ValidationIssue
}

var serviceValidators []ServiceValidator

func RegisterServiceValidator(v ServiceValidator) {
	serviceValidators = append(serviceValidators, v)
}

func ResetServiceValidators() {
	serviceValidators = nil
}

var validProtocols = []string{"http", "ssh", "tcp", "mcp", "telnet"}

var validCommandPlugins = []string{"", "LLMHoneypot", "MazeHoneypot"}

var validCommandPluginsDisplay = []string{"(none)", "LLMHoneypot", "MazeHoneypot"}

func Validate(services []BeelzebubServiceConfiguration, parseIssues []ValidationIssue) ValidateResult {
	resultMap := make(map[string]*ValidationResult)

	for _, issue := range parseIssues {
		r, ok := resultMap[issue.Filename]
		if !ok {
			r = &ValidationResult{Filename: issue.Filename}
			resultMap[issue.Filename] = r
		}
		r.Issues = append(r.Issues, issue)
	}

	for i := range services {
		filename := services[i].Filename
		r, ok := resultMap[filename]
		if !ok {
			r = &ValidationResult{Filename: filename}
			resultMap[filename] = r
		}

		protocol := services[i].Protocol
		if !slices.Contains(validProtocols, protocol) {
			r.Issues = append(r.Issues, ValidationIssue{
				Level:   LevelError,
				Message: fmt.Sprintf("invalid protocol %q, valid: %s", protocol, strings.Join(validProtocols, ", ")),
			})
		}

		address := strings.TrimSpace(services[i].Address)
		if address == "" {
			r.Issues = append(r.Issues, ValidationIssue{
				Level:   LevelError,
				Message: "address is empty",
			})
		} else if !strings.Contains(address, "/") {
			lastColon := strings.LastIndex(address, ":")
			if lastColon == -1 {
				r.Issues = append(r.Issues, ValidationIssue{
					Level:   LevelWarning,
					Message: fmt.Sprintf("address %q has invalid port format or port out of range (1-65535)", address),
				})
			} else {
				portStr := address[lastColon+1:]
				port, err := strconv.Atoi(portStr)
				if err != nil || port < 1 || port > 65535 {
					r.Issues = append(r.Issues, ValidationIssue{
						Level:   LevelWarning,
						Message: fmt.Sprintf("address %q has invalid port format or port out of range (1-65535)", address),
					})
				}
			}
		}

		for j, cmd := range services[i].Commands {
			if cmd.RegexStr == "" {
				r.Issues = append(r.Issues, ValidationIssue{
					Level:   LevelError,
					Message: fmt.Sprintf("command[%d] has empty regex", j),
				})
			}
			if !slices.Contains(validCommandPlugins, cmd.Plugin) {
				r.Issues = append(r.Issues, ValidationIssue{
					Level:   LevelError,
					Message: fmt.Sprintf("command[%d] has invalid plugin %q, valid: %s", j, cmd.Plugin, strings.Join(validCommandPluginsDisplay, ", ")),
				})
			}
			if cmd.Handler == "" && cmd.Plugin == "" {
				r.Issues = append(r.Issues, ValidationIssue{
					Level:   LevelWarning,
					Message: fmt.Sprintf("command[%d] has empty handler and no plugin — matched requests will produce no output", j),
				})
			}
			for _, header := range cmd.Headers {
				if !strings.Contains(header, ":") {
					r.Issues = append(r.Issues, ValidationIssue{
						Level:   LevelWarning,
						Message: fmt.Sprintf("command[%d].headers has malformed entry (no colon): %q", j, header),
					})
				}
			}
		}

		fb := services[i].FallbackCommand
		if fb.Handler != "" || fb.Plugin != "" {
			if fb.RegexStr != "" {
				if _, err := regexp.Compile(fb.RegexStr); err != nil {
					r.Issues = append(r.Issues, ValidationIssue{
						Level:   LevelError,
						Message: fmt.Sprintf("fallbackCommand has invalid regex: %v", err),
					})
				}
			}
			if !slices.Contains(validCommandPlugins, fb.Plugin) {
				r.Issues = append(r.Issues, ValidationIssue{
					Level:   LevelError,
					Message: fmt.Sprintf("fallbackCommand has invalid plugin %q, valid: %s", fb.Plugin, strings.Join(validCommandPluginsDisplay, ", ")),
				})
			}
		}

		if services[i].Plugin.OpenAISecretKey != "" {
			r.Issues = append(r.Issues, ValidationIssue{
				Level:   LevelWarning,
				Message: "openAISecretKey is set inline, consider using env var OPEN_AI_SECRET_KEY",
			})
		}

		if services[i].DeadlineTimeoutSeconds == 0 && len(services[i].Commands) > 0 {
			r.Issues = append(r.Issues, ValidationIssue{
				Level:   LevelWarning,
				Message: "deadlineTimeoutSeconds is not set, connections may be closed immediately",
			})
		}

		for _, v := range serviceValidators {
			issues := v.Validate(services[i])
			r.Issues = append(r.Issues, issues...)
		}
	}

	collisionMap := make(map[string][]int)
	for i, svc := range services {
		key := svc.Protocol + ":" + svc.Address
		collisionMap[key] = append(collisionMap[key], i)
	}

	for _, indices := range collisionMap {
		if len(indices) > 1 {
			for _, idx := range indices {
				svc := services[idx]
				r := resultMap[svc.Filename]
				r.Issues = append(r.Issues, ValidationIssue{
					Level:   LevelError,
					Message: fmt.Sprintf("address %s:%s is used by multiple services", svc.Protocol, svc.Address),
				})
			}
		}
	}

	var results []ValidationResult
	for _, r := range resultMap {
		results = append(results, *r)
	}

	var totalErrors, totalWarnings int
	for _, r := range results {
		for _, issue := range r.Issues {
			switch issue.Level {
			case LevelError:
				totalErrors++
			case LevelWarning:
				totalWarnings++
			}
		}
	}

	return ValidateResult{
		Results:       results,
		TotalErrors:   totalErrors,
		TotalWarnings: totalWarnings,
	}
}

func (r ValidateResult) Print() {
	sort.Slice(r.Results, func(i, j int) bool {
		return r.Results[i].Filename < r.Results[j].Filename
	})

	for _, result := range r.Results {
		hasErrors := false
		for _, issue := range result.Issues {
			if issue.Level == LevelError {
				hasErrors = true
				break
			}
		}

		if len(result.Issues) == 0 {
			fmt.Printf("OK %s\n", result.Filename)
		} else if hasErrors {
			fmt.Printf("FAIL %s\n", result.Filename)
			for _, issue := range result.Issues {
				fmt.Printf("  %s: %s\n", strings.ToUpper(issue.Level), issue.Message)
			}
		} else {
			fmt.Printf("WARN %s\n", result.Filename)
			for _, issue := range result.Issues {
				fmt.Printf("  %s: %s\n", strings.ToUpper(issue.Level), issue.Message)
			}
		}
	}

	fmt.Printf("\n%d errors, %d warnings\n", r.TotalErrors, r.TotalWarnings)
}

func (r ValidateResult) ExitCode() int {
	if r.TotalErrors > 0 {
		return 1
	}
	return 0
}

func ValidateCore(config *BeelzebubCoreConfigurations, filename string) ValidateResult {
	var issues []ValidationIssue

	if config.Core.Tracings.RabbitMQ.Enabled && config.Core.Tracings.RabbitMQ.URI == "" {
		issues = append(issues, ValidationIssue{
			Level:    LevelError,
			Message:  "rabbitMQ is enabled but URI is empty",
			Filename: filename,
		})
	}

	if config.Core.BeelzebubCloud.Enabled {
		if config.Core.BeelzebubCloud.URI == "" {
			issues = append(issues, ValidationIssue{
				Level:    LevelError,
				Message:  "beelzebub-cloud is enabled but URI is empty",
				Filename: filename,
			})
		}
		if config.Core.BeelzebubCloud.AuthToken == "" {
			issues = append(issues, ValidationIssue{
				Level:    LevelError,
				Message:  "beelzebub-cloud is enabled but auth-token is empty",
				Filename: filename,
			})
		}
	}

	if config.Core.Prometheus != (Prometheus{}) {
		if config.Core.Prometheus.Path == "" {
			issues = append(issues, ValidationIssue{
				Level:    LevelError,
				Message:  "prometheus is configured but path is empty",
				Filename: filename,
			})
		} else if !strings.HasPrefix(config.Core.Prometheus.Path, "/") {
			issues = append(issues, ValidationIssue{
				Level:    LevelError,
				Message:  fmt.Sprintf("prometheus path %q must start with /", config.Core.Prometheus.Path),
				Filename: filename,
			})
		}
		if config.Core.Prometheus.Port == "" {
			issues = append(issues, ValidationIssue{
				Level:    LevelError,
				Message:  "prometheus is configured but port is empty",
				Filename: filename,
			})
		}
	}

	if config.Core.Logging.LogsPath != "" {
		parentDir := filepath.Dir(config.Core.Logging.LogsPath)
		if _, err := os.Stat(parentDir); os.IsNotExist(err) {
			issues = append(issues, ValidationIssue{
				Level:    LevelWarning,
				Message:  fmt.Sprintf("logs path %q: parent directory %q does not exist", config.Core.Logging.LogsPath, parentDir),
				Filename: filename,
			})
		}
	}

	var totalErrors int
	for _, issue := range issues {
		if issue.Level == LevelError {
			totalErrors++
		}
	}

	return ValidateResult{
		Results: []ValidationResult{
			{Filename: filename, Issues: issues},
		},
		TotalErrors:   totalErrors,
		TotalWarnings: len(issues) - totalErrors,
	}
}
