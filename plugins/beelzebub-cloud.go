package plugins

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/mariocandela/beelzebub/v3/parser"
	"github.com/mariocandela/beelzebub/v3/tracer"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type beelzebubCloud struct {
	URI       string
	AuthToken string
	client    *resty.Client
}

type HoneypotConfigResponseDTO struct {
	ID            string `json:"id"`
	Config        string `json:"config"`
	TokenID       string `json:"tokenId"`
	LastUpdatedOn string `json:"lastUpdatedOn"`
}

func InitBeelzebubCloud(uri, authToken string) *beelzebubCloud {
	return &beelzebubCloud{
		URI:       uri,
		AuthToken: authToken,
		client:    resty.New(),
	}
}

func (beelzebubCloud *beelzebubCloud) SendEvent(event tracer.Event) (bool, error) {
	requestJson, err := json.Marshal(event)
	if err != nil {
		return false, err
	}

	if beelzebubCloud.AuthToken == "" {
		return false, errors.New("authToken is empty")
	}

	response, err := beelzebubCloud.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(requestJson).
		SetHeader("Authorization", beelzebubCloud.AuthToken).
		SetResult(&tracer.Event{}).
		Post(fmt.Sprintf("%s/events", beelzebubCloud.URI))

	log.Debug(response)

	if err != nil {
		return false, err
	}

	return response.StatusCode() == 200, nil
}

func (beelzebubCloud *beelzebubCloud) GetHoneypotsConfigurations() ([]parser.BeelzebubServiceConfiguration, error) {
	if beelzebubCloud.AuthToken == "" {
		return nil, errors.New("authToken is empty")
	}

	response, err := beelzebubCloud.client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", beelzebubCloud.AuthToken).
		SetResult([]HoneypotConfigResponseDTO{}).
		Get(fmt.Sprintf("%s/honeypots", beelzebubCloud.URI))

	if err != nil {
		return nil, err
	}

	if response.StatusCode() != 200 {
		return nil, errors.New(fmt.Sprintf("Response code: %v, error: %s", response.StatusCode(), string(response.Body())))
	}

	var honeypotsConfig []HoneypotConfigResponseDTO

	if err = json.Unmarshal(response.Body(), &honeypotsConfig); err != nil {
		return nil, err
	}

	var servicesConfiguration = make([]parser.BeelzebubServiceConfiguration, 0)

	for _, honeypotConfig := range honeypotsConfig {
		var honeypotsConfig parser.BeelzebubServiceConfiguration

		if err = yaml.Unmarshal([]byte(honeypotConfig.Config), &honeypotsConfig); err != nil {
			return nil, err
		}
		if err := honeypotsConfig.CompileCommandRegex(); err != nil {
			return nil, fmt.Errorf("unable to load service config from cloud: invalid regex: %v", err)
		}
		servicesConfiguration = append(servicesConfiguration, honeypotsConfig)
	}

	log.Debug(servicesConfiguration)

	return servicesConfiguration, nil
}
