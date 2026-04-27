package HTTP

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/beelzebub-labs/beelzebub/v3/internal/parser"
	"github.com/beelzebub-labs/beelzebub/v3/internal/tracer"
	"github.com/stretchr/testify/assert"
)

type mockTracer struct {
	events []tracer.Event
}

func (m *mockTracer) TraceEvent(event tracer.Event) {
	m.events = append(m.events, event)
}

func TestMapHeaderToString_Empty(t *testing.T) {
	result := mapHeaderToString(http.Header{})
	assert.Equal(t, "", result)
}

func TestMapHeaderToString_SingleHeader(t *testing.T) {
	headers := http.Header{}
	headers.Set("Content-Type", "application/json")

	result := mapHeaderToString(headers)

	assert.Contains(t, result, "Content-Type")
	assert.Contains(t, result, "application/json")
}

func TestMapHeaderToString_MultipleHeaders(t *testing.T) {
	headers := http.Header{}
	headers.Set("Content-Type", "text/html")
	headers.Set("X-Custom", "value")

	result := mapHeaderToString(headers)

	assert.Contains(t, result, "Content-Type")
	assert.Contains(t, result, "X-Custom")
}

func TestMapCookiesToString_Empty(t *testing.T) {
	result := mapCookiesToString([]*http.Cookie{})
	assert.Equal(t, "", result)
}

func TestMapCookiesToString_SingleCookie(t *testing.T) {
	cookies := []*http.Cookie{
		{Name: "session", Value: "abc123"},
	}

	result := mapCookiesToString(cookies)

	assert.Contains(t, result, "session")
	assert.Contains(t, result, "abc123")
}

func TestMapCookiesToString_MultipleCookies(t *testing.T) {
	cookies := []*http.Cookie{
		{Name: "session", Value: "abc123"},
		{Name: "user", Value: "john"},
	}

	result := mapCookiesToString(cookies)

	assert.Contains(t, result, "session")
	assert.Contains(t, result, "user")
}

func TestSetResponseHeaders_ValidStatusCode(t *testing.T) {
	w := httptest.NewRecorder()
	setResponseHeaders(w, []string{"Content-Type: application/json"}, http.StatusOK)

	assert.Equal(t, http.StatusOK, w.Code)
	// The implementation splits on ":" and preserves the space, so the value has a leading space
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
}

func TestSetResponseHeaders_InvalidStatusCode(t *testing.T) {
	w := httptest.NewRecorder()
	setResponseHeaders(w, []string{}, 999)

	// Unknown status code: WriteHeader should not be called, default stays 200
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSetResponseHeaders_NoHeaders(t *testing.T) {
	w := httptest.NewRecorder()
	setResponseHeaders(w, []string{}, http.StatusNotFound)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestSetResponseHeaders_HeaderWithoutColon(t *testing.T) {
	w := httptest.NewRecorder()
	setResponseHeaders(w, []string{"InvalidHeader"}, http.StatusOK)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTraceRequest_HTTP(t *testing.T) {
	mt := &mockTracer{}
	req := httptest.NewRequest(http.MethodGet, "/test", strings.NewReader("body"))
	req.Header.Set("User-Agent", "TestAgent/1.0")
	req.RemoteAddr = "127.0.0.1:12345"

	cmd := parser.Command{Name: "test-handler"}
	traceRequest(req, mt, cmd, "test-honeypot", "body")

	assert.Len(t, mt.events, 1)
	event := mt.events[0]
	assert.Equal(t, "HTTP New request", event.Msg)
	assert.Equal(t, tracer.HTTP.String(), event.Protocol)
	assert.Equal(t, "test-honeypot", event.Description)
	assert.Equal(t, "127.0.0.1", event.SourceIp)
	assert.Equal(t, "12345", event.SourcePort)
	assert.Equal(t, "test-handler", event.Handler)
}

func TestTraceRequest_WithCookiesAndHeaders(t *testing.T) {
	mt := &mockTracer{}
	req := httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(`{"user":"admin"}`))
	req.Header.Set("X-Forwarded-For", "10.0.0.1")
	req.AddCookie(&http.Cookie{Name: "session", Value: "xyz"})
	req.RemoteAddr = "192.168.1.1:54321"

	traceRequest(req, mt, parser.Command{}, "login-honeypot", `{"user":"admin"}`)

	assert.Len(t, mt.events, 1)
	event := mt.events[0]
	assert.Contains(t, event.Headers, "X-Forwarded-For")
	assert.Contains(t, event.Cookies, "session")
	assert.Equal(t, "192.168.1.1", event.SourceIp)
}

func TestBuildHTTPResponse_StaticHandler(t *testing.T) {
	mt := &mockTracer{}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "127.0.0.1:1234"

	servConf := parser.BeelzebubServiceConfiguration{Description: "test"}
	cmd := parser.Command{
		Handler:    "Hello World",
		StatusCode: http.StatusOK,
		Headers:    []string{"Content-Type: text/plain"},
	}

	resp, err := buildHTTPResponse(servConf, mt, cmd, req)

	assert.NoError(t, err)
	assert.Equal(t, "Hello World", resp.Body)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestBuildHTTPResponse_UnknownPlugin(t *testing.T) {
	mt := &mockTracer{}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "127.0.0.1:1234"

	servConf := parser.BeelzebubServiceConfiguration{Description: "test"}
	cmd := parser.Command{
		Plugin: "non-existent-plugin-xyz",
	}

	resp, err := buildHTTPResponse(servConf, mt, cmd, req)

	assert.NoError(t, err)
	// Falls through to unknown plugin branch; body stays empty
	assert.Equal(t, "", resp.Body)
}
