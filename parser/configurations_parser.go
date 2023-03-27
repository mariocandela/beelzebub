package parser

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type BeelzebubCoreConfigurations struct {
	Core struct {
		Logging    Logging    `yaml:"logging"`
		Tracing    Tracing    `yaml:"tracing"`
		Prometheus Prometheus `yaml:"prometheus"`
	}
}

type Logging struct {
	Debug               bool   `yaml:"debug"`
	DebugReportCaller   bool   `yaml:"debugReportCaller"`
	LogDisableTimestamp bool   `yaml:"logDisableTimestamp"`
	LogsPath            string `yaml:"logsPath,omitempty"`
}

type Tracing struct {
	RabbitMQEnabled bool   `yaml:"rabbitMQEnabled"`
	RabbitMQURI     string `yaml:"rabbitMQURI"`
}

type Prometheus struct {
	Path string `yaml:"path"`
	Port string `yaml:"port"`
}

type Plugin struct {
	OpenAPIChatGPTSecretKey string `yaml:"openAPIChatGPTSecretKey"`
}

type BeelzebubServiceConfiguration struct {
	ApiVersion             string    `yaml:"apiVersion"`
	Protocol               string    `yaml:"protocol"`
	Address                string    `yaml:"address"`
	Commands               []Command `yaml:"commands"`
	ServerVersion          string    `yaml:"serverVersion"`
	ServerName             string    `yaml:"serverName"`
	DeadlineTimeoutSeconds int       `yaml:"deadlineTimeoutSeconds"`
	PasswordRegex          string    `yaml:"passwordRegex"`
	Description            string    `yaml:"description"`
	Banner                 string    `yaml:"banner"`
	Plugin                 Plugin    `yaml:"plugin"`
}

type Command struct {
	Regex      string   `yaml:"regex"`
	Handler    string   `yaml:"handler"`
	Headers    []string `yaml:"headers"`
	StatusCode int      `yaml:"statusCode"`
	Plugin     string   `yaml:"plugin"`
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

func (bp configurationsParser) ReadConfigurationsCore() (*BeelzebubCoreConfigurations, error) {
	buf, err := bp.readFileBytesByFilePathDependency(bp.configurationsCorePath)
	if err != nil {
		return nil, fmt.Errorf("in file %s: %v", bp.configurationsCorePath, err)
	}

	beelzebubConfiguration := &BeelzebubCoreConfigurations{}
	err = yaml.Unmarshal(buf, beelzebubConfiguration)
	if err != nil {
		return nil, fmt.Errorf("in file %s: %v", bp.configurationsCorePath, err)
	}

	return beelzebubConfiguration, nil
}

func (bp configurationsParser) ReadConfigurationsServices() ([]BeelzebubServiceConfiguration, error) {
	services, err := bp.gelAllFilesNameByDirNameDependency(bp.configurationsServicesDirectory)
	if err != nil {
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
		log.Debug(beelzebubServiceConfiguration)
		servicesConfiguration = append(servicesConfiguration, *beelzebubServiceConfiguration)
	}

	return servicesConfiguration, nil
}

func gelAllFilesNameByDirName(dirName string) ([]string, error) {
	var filesName []string
	files, err := ioutil.ReadDir(dirName)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		filesName = append(filesName, file.Name())
	}
	return filesName, nil
}

func readFileBytesByFilePath(filePath string) ([]byte, error) {
	return os.ReadFile(filePath)
}
