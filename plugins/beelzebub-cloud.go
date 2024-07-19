package plugins

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/mariocandela/beelzebub/v3/parser"
	"github.com/mariocandela/beelzebub/v3/tracer"
	log "github.com/sirupsen/logrus"
)

type beelzebubCloud struct {
	URI       string
	AuthToken string
	client    *resty.Client
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
		SetResult(&gptResponse{}).
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
		SetResult(&gptResponse{}).
		Get(fmt.Sprintf("%s/honeypots", beelzebubCloud.URI))

	log.Debug(response)

	if err != nil {
		return nil, err
	}

	if response.StatusCode() != 200 {
		return nil, errors.New(fmt.Sprintf("Response code: %v, error: %s", response.StatusCode(), string(response.Body())))
	}

	var honeypotsConfiguration []parser.BeelzebubServiceConfiguration

	if err = json.Unmarshal(response.Body(), &honeypotsConfiguration); err != nil {
		return nil, err
	}

	return honeypotsConfiguration, nil
}
