// Package parser is responsible for parsing the configurations of the core and honeypot service
package parser

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// BeelzebubCoreConfigurations is the struct that contains the configurations of the core
type BeelzebubCoreConfigurations struct {
	Core struct {
		Logging        Logging        `yaml:"logging"`
		Tracings       Tracings       `yaml:"tracings"`
		Prometheus     Prometheus     `yaml:"prometheus"`
		BeelzebubCloud BeelzebubCloud `yaml:"beelzebub-cloud"`
	}
}

// Logging is the struct that contains the configurations of the logging
type Logging struct {
	Debug               bool   `yaml:"debug"`
	DebugReportCaller   bool   `yaml:"debugReportCaller"`
	LogDisableTimestamp bool   `yaml:"logDisableTimestamp"`
	LogsPath            string `yaml:"logsPath,omitempty"`
}

// Tracings is the struct that contains the configurations of the tracings
type Tracings struct {
	RabbitMQ `yaml:"rabbit-mq"`
}

type BeelzebubCloud struct {
	Enabled   bool   `yaml:"enabled"`
	URI       string `yaml:"uri"`
	AuthToken string `yaml:"auth-token"`
}
type RabbitMQ struct {
	Enabled bool   `yaml:"enabled"`
	URI     string `yaml:"uri"`
}
type Prometheus struct {
	Path string `yaml:"path"`
	Port string `yaml:"port"`
}

type Plugin struct {
	OpenAISecretKey         string `yaml:"openAISecretKey"`
	Host                    string `yaml:"host"`
	LLMModel                string `yaml:"llmModel"`
	LLMProvider             string `yaml:"llmProvider"`
	Prompt                  string `yaml:"prompt"`
	InputValidationEnabled  bool   `yaml:"inputValidationEnabled"`
	InputValidationPrompt   string `yaml:"inputValidationPrompt"`
	OutputValidationEnabled bool   `yaml:"outputValidationEnabled"`
	OutputValidationPrompt  string `yaml:"outputValidationPrompt"`
	RateLimitEnabled        bool   `yaml:"rateLimitEnabled"`
	RateLimitRequests       int    `yaml:"rateLimitRequests"`
	RateLimitWindowSeconds  int    `yaml:"rateLimitWindowSeconds"`
}

// BeelzebubServiceConfiguration is the struct that contains the configurations of the honeypot service
type BeelzebubServiceConfiguration struct {
	ApiVersion             string    `yaml:"apiVersion"`
	Protocol               string    `yaml:"protocol"`
	Address                string    `yaml:"address"`
	Commands               []Command `yaml:"commands"`
	Tools                  []Tool    `yaml:"tools"`
	FallbackCommand        Command   `yaml:"fallbackCommand"`
	ServerVersion          string    `yaml:"serverVersion"`
	ServerName             string    `yaml:"serverName"`
	DeadlineTimeoutSeconds int       `yaml:"deadlineTimeoutSeconds"`
	PasswordRegex          string    `yaml:"passwordRegex"`
	Description            string    `yaml:"description"`
	Banner                 string    `yaml:"banner"`
	Plugin                 Plugin    `yaml:"plugin"`
	TLSCertPath            string    `yaml:"tlsCertPath"`
	TLSKeyPath             string    `yaml:"tlsKeyPath"`
	// TrustedProxies is a list of CIDRs (or bare IPs) of upstream proxies whose
	// X-Forwarded-For / X-Real-IP headers can be trusted. When empty, those
	// headers are ignored and the immediate TCP peer is used as source IP.
	TrustedProxies     []string     `yaml:"trustedProxies,omitempty" json:",omitempty"`
	TrustedProxiesNets []*net.IPNet `yaml:"-" json:"-"`
}

func (bsc BeelzebubServiceConfiguration) HashCode() (string, error) {
	data, err := json.Marshal(bsc)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}

// Command is the struct that contains the configurations of the commands
type Command struct {
	RegexStr   string         `yaml:"regex"`
	Regex      *regexp.Regexp `yaml:"-"` // This field is parsed, not stored in the config itself.
	Handler    string         `yaml:"handler"`
	Headers    []string       `yaml:"headers"`
	StatusCode int            `yaml:"statusCode"`
	Plugin     string         `yaml:"plugin"`
	Name       string         `yaml:"name"`
}

// Tool is the struct that contains the configurations of the MCP Honeypot
type Tool struct {
	Name        string           `yaml:"name" json:"Name"`
	Description string           `yaml:"description" json:"Description"`
	Params      []Param          `yaml:"params" json:"Params"`
	Handler     string           `yaml:"handler" json:"Handler"`
	Annotations *ToolAnnotations `yaml:"annotations,omitempty" json:"Annotations,omitempty"`
}

// ToolAnnotations contains MCP tool annotation hints for LLM clients
type ToolAnnotations struct {
	Title           string `yaml:"title,omitempty" json:"Title,omitempty"`
	ReadOnlyHint    *bool  `yaml:"readOnlyHint,omitempty" json:"ReadOnlyHint,omitempty"`
	DestructiveHint *bool  `yaml:"destructiveHint,omitempty" json:"DestructiveHint,omitempty"`
	IdempotentHint  *bool  `yaml:"idempotentHint,omitempty" json:"IdempotentHint,omitempty"`
	OpenWorldHint   *bool  `yaml:"openWorldHint,omitempty" json:"OpenWorldHint,omitempty"`
}

// Param is the struct that contains the configurations of the parameters of the tools
type Param struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

type configurationsParser struct {
	configurationsCorePath             string
	configurationsServicesDirectory    string
	readFileBytesByFilePathDependency  ReadFileBytesByFilePath
	gelAllFilesNameByDirNameDependency GelAllFilesNameByDirName
}

type ReadFileBytesByFilePath func(filePath string) ([]byte, error)

type GelAllFilesNameByDirName func(dirName string) ([]string, error)

// Init Parser, return a configurationsParser and use the D.I. Pattern to inject the dependencies
func Init(configurationsCorePath, configurationsServicesDirectory string) *configurationsParser {
	return &configurationsParser{
		configurationsCorePath:             configurationsCorePath,
		configurationsServicesDirectory:    configurationsServicesDirectory,
		readFileBytesByFilePathDependency:  readFileBytesByFilePath,
		gelAllFilesNameByDirNameDependency: gelAllFilesNameByDirName,
	}
}

// ReadConfigurationsCore is the method that reads the configurations of the core from files.
// If the file does not exist, a default empty configuration is used.
// Environment variables always override file values (see applyEnvOverrides).
func (bp configurationsParser) ReadConfigurationsCore() (*BeelzebubCoreConfigurations, error) {
	buf, err := bp.readFileBytesByFilePathDependency(bp.configurationsCorePath)
	if err != nil {
		if !isNotFound(err) {
			return nil, fmt.Errorf("in file %s: %v", bp.configurationsCorePath, err)
		}
		log.Debug("Core config file not found, falling back to environment variables")
		buf = []byte{}
	}

	beelzebubConfiguration := &BeelzebubCoreConfigurations{}
	if err = yaml.Unmarshal(buf, beelzebubConfiguration); err != nil {
		return nil, fmt.Errorf("in file %s: %v", bp.configurationsCorePath, err)
	}

	applyEnvOverrides(beelzebubConfiguration)
	return beelzebubConfiguration, nil
}

// applyEnvOverrides overrides configuration fields with environment variable values when set.
// Supported variables:
//
//	BEELZEBUB_LOGGING_DEBUG, BEELZEBUB_LOGGING_DEBUG_REPORT_CALLER,
//	BEELZEBUB_LOGGING_LOG_DISABLE_TIMESTAMP, BEELZEBUB_LOGGING_LOGS_PATH,
//	BEELZEBUB_RABBITMQ_ENABLED, BEELZEBUB_RABBITMQ_URI,
//	BEELZEBUB_PROMETHEUS_PATH, BEELZEBUB_PROMETHEUS_PORT,
//	BEELZEBUB_CLOUD_ENABLED, BEELZEBUB_CLOUD_URI, BEELZEBUB_CLOUD_AUTH_TOKEN
func applyEnvOverrides(cfg *BeelzebubCoreConfigurations) {
	if v := os.Getenv("BEELZEBUB_LOGGING_DEBUG"); v != "" {
		cfg.Core.Logging.Debug = parseBool(v)
	}
	if v := os.Getenv("BEELZEBUB_LOGGING_DEBUG_REPORT_CALLER"); v != "" {
		cfg.Core.Logging.DebugReportCaller = parseBool(v)
	}
	if v := os.Getenv("BEELZEBUB_LOGGING_LOG_DISABLE_TIMESTAMP"); v != "" {
		cfg.Core.Logging.LogDisableTimestamp = parseBool(v)
	}
	if v := os.Getenv("BEELZEBUB_LOGGING_LOGS_PATH"); v != "" {
		cfg.Core.Logging.LogsPath = v
	}
	if v := os.Getenv("BEELZEBUB_RABBITMQ_ENABLED"); v != "" {
		cfg.Core.Tracings.RabbitMQ.Enabled = parseBool(v)
	}
	if v := os.Getenv("BEELZEBUB_RABBITMQ_URI"); v != "" {
		cfg.Core.Tracings.RabbitMQ.URI = v
	}
	if v := os.Getenv("BEELZEBUB_PROMETHEUS_PATH"); v != "" {
		cfg.Core.Prometheus.Path = v
	}
	if v := os.Getenv("BEELZEBUB_PROMETHEUS_PORT"); v != "" {
		cfg.Core.Prometheus.Port = v
	}
	if v := os.Getenv("BEELZEBUB_CLOUD_ENABLED"); v != "" {
		cfg.Core.BeelzebubCloud.Enabled = parseBool(v)
	}
	if v := os.Getenv("BEELZEBUB_CLOUD_URI"); v != "" {
		cfg.Core.BeelzebubCloud.URI = v
	}
	if v := os.Getenv("BEELZEBUB_CLOUD_AUTH_TOKEN"); v != "" {
		cfg.Core.BeelzebubCloud.AuthToken = v
	}
}

func parseBool(v string) bool {
	b, _ := strconv.ParseBool(v)
	return b
}

func isNotFound(err error) bool {
	return os.IsNotExist(err)
}

// ReadConfigurationsServices is the method that reads the configurations of the honeypot services.
// If the BEELZEBUB_SERVICES_CONFIG environment variable is set (JSON array), it is used directly.
// Otherwise, service YAML files are loaded from the configured directory (existing behaviour).
func (bp configurationsParser) ReadConfigurationsServices() ([]BeelzebubServiceConfiguration, error) {
	if envConfig := os.Getenv("BEELZEBUB_SERVICES_CONFIG"); envConfig != "" {
		return parseServicesFromEnv(envConfig)
	}

	services, err := bp.gelAllFilesNameByDirNameDependency(bp.configurationsServicesDirectory)

	if err != nil {
		if isNotFound(err) {
			log.Warnf("Services config directory %q not found, falling back to empty configuration", bp.configurationsServicesDirectory)
			return []BeelzebubServiceConfiguration{}, nil
		}
		return nil, fmt.Errorf("in directory %s: %v", bp.configurationsServicesDirectory, err)
	}

	var servicesConfiguration []BeelzebubServiceConfiguration

	for _, servicesName := range services {
		filePath := filepath.Join(bp.configurationsServicesDirectory, servicesName)
		buf, err := bp.readFileBytesByFilePathDependency(filePath)

		if err != nil {
			return nil, fmt.Errorf("in file %s: %v", filePath, err)
		}

		beelzebubServiceConfiguration := &BeelzebubServiceConfiguration{}
		err = yaml.Unmarshal(buf, beelzebubServiceConfiguration)

		if err != nil {
			return nil, fmt.Errorf("in file %s: %v", filePath, err)
		}

		if beelzebubServiceConfiguration.Plugin.RateLimitEnabled {
			if beelzebubServiceConfiguration.Plugin.RateLimitRequests <= 0 ||
				beelzebubServiceConfiguration.Plugin.RateLimitWindowSeconds <= 0 {
				return nil, fmt.Errorf("in file %s: invalid rate limiting config: rateLimitRequests and rateLimitWindowSeconds must be > 0", filePath)
			}
		}

		log.Debug(beelzebubServiceConfiguration)

		if err := beelzebubServiceConfiguration.CompileCommandRegex(); err != nil {
			return nil, fmt.Errorf("in file %s: invalid regex: %v", filePath, err)
		}

		if err := beelzebubServiceConfiguration.CompileTrustedProxies(); err != nil {
			return nil, fmt.Errorf("in file %s: %v", filePath, err)
		}

		servicesConfiguration = append(servicesConfiguration, *beelzebubServiceConfiguration)
	}

	return servicesConfiguration, nil
}

// parseServicesFromEnv parses a JSON array of BeelzebubServiceConfiguration from
// the BEELZEBUB_SERVICES_CONFIG environment variable.
func parseServicesFromEnv(jsonStr string) ([]BeelzebubServiceConfiguration, error) {
	var servicesConfiguration []BeelzebubServiceConfiguration
	if err := json.Unmarshal([]byte(jsonStr), &servicesConfiguration); err != nil {
		return nil, fmt.Errorf("invalid BEELZEBUB_SERVICES_CONFIG: %v", err)
	}

	for i := range servicesConfiguration {
		svc := &servicesConfiguration[i]

		if svc.Plugin.RateLimitEnabled {
			if svc.Plugin.RateLimitRequests <= 0 || svc.Plugin.RateLimitWindowSeconds <= 0 {
				return nil, fmt.Errorf("invalid rate limiting config in BEELZEBUB_SERVICES_CONFIG[%d]: rateLimitRequests and rateLimitWindowSeconds must be > 0", i)
			}
		}

		if err := svc.CompileCommandRegex(); err != nil {
			return nil, fmt.Errorf("invalid regex in BEELZEBUB_SERVICES_CONFIG[%d]: %v", i, err)
		}

		if err := svc.CompileTrustedProxies(); err != nil {
			return nil, fmt.Errorf("in BEELZEBUB_SERVICES_CONFIG[%d]: %v", i, err)
		}
	}

	return servicesConfiguration, nil
}

// CompileCommandRegex is the method that compiles the regular expression for each configured Command.
func (c *BeelzebubServiceConfiguration) CompileCommandRegex() error {
	for i, command := range c.Commands {
		if command.RegexStr != "" {
			rex, err := regexp.Compile(command.RegexStr)
			if err != nil {
				return err
			}
			c.Commands[i].Regex = rex
		}
	}
	return nil
}

// CompileTrustedProxies parses the TrustedProxies entries (CIDRs or bare IPs)
// into net.IPNet values stored in TrustedProxiesNets. Bare IPs are treated as
// /32 (IPv4) or /128 (IPv6).
func (c *BeelzebubServiceConfiguration) CompileTrustedProxies() error {
	nets := make([]*net.IPNet, 0, len(c.TrustedProxies))
	for _, entry := range c.TrustedProxies {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		if !strings.Contains(entry, "/") {
			ip := net.ParseIP(entry)
			if ip == nil {
				return fmt.Errorf("invalid trustedProxies entry %q", entry)
			}
			if ip.To4() != nil {
				entry += "/32"
			} else {
				entry += "/128"
			}
		}
		_, n, err := net.ParseCIDR(entry)
		if err != nil {
			return fmt.Errorf("invalid trustedProxies entry %q: %v", entry, err)
		}
		nets = append(nets, n)
	}
	c.TrustedProxiesNets = nets
	return nil
}

func gelAllFilesNameByDirName(dirName string) ([]string, error) {
	files, err := os.ReadDir(dirName)
	if err != nil {
		return nil, err
	}

	var filesName []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".yaml") {
			filesName = append(filesName, file.Name())
		}
	}
	return filesName, nil
}

func readFileBytesByFilePath(filePath string) ([]byte, error) {
	return os.ReadFile(filePath)
}
