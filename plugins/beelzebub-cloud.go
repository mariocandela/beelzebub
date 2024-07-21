package plugins

import (
	"encoding/json"
	"errors"
	"github.com/go-resty/resty/v2"
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
		Post(beelzebubCloud.URI)

	log.Debug(response)

	if err != nil {
		return false, err
	}

	return response.StatusCode() == 200, nil
}
