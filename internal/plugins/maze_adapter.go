package plugins

import (
	"net/http"

	"github.com/mariocandela/beelzebub/v3/pkg/plugin"
)

// mazePlugin is the registry adapter for MazeHoneypot.
// It implements plugin.HTTPPlugin so the HTTP strategy can dispatch to it
// via the registry instead of hardcoded string checks.
type mazePlugin struct{}

func (m *mazePlugin) Metadata() plugin.Metadata {
	return plugin.Metadata{
		Name:        MazePluginName,
		Description: "Infinite deterministic directory maze — generates realistic Apache-style directory listings",
		Version:     "1.0.0",
		Author:      "beelzebub",
	}
}

func (m *mazePlugin) HandleHTTP(r *http.Request) plugin.HTTPResponse {
	maze := &MazeHoneypot{}

	// Config values (ServerVersion, ServerName) are injected at dispatch time
	// by the HTTP strategy using the service configuration, so they are not
	// available here. The HTTP strategy sets them on a fresh MazeHoneypot
	// directly when it needs configuration-aware behaviour.
	resp := maze.HandleRequest(r)
	headers := make(map[string]string, len(resp.Headers))
	for k, v := range resp.Headers {
		headers[k] = v
	}
	return plugin.HTTPResponse{
		StatusCode:  resp.StatusCode,
		Body:        resp.Body,
		Headers:     headers,
		ContentType: resp.ContentType,
	}
}

func init() {
	plugin.Register(&mazePlugin{})
}
