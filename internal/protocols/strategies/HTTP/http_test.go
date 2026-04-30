package HTTP

import (
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/beelzebub-labs/beelzebub/v3/internal/parser"
	"github.com/beelzebub-labs/beelzebub/v3/internal/tracer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mustCIDRs(t *testing.T, cidrs ...string) []*net.IPNet {
	t.Helper()
	nets := make([]*net.IPNet, 0, len(cidrs))
	for _, c := range cidrs {
		_, n, err := net.ParseCIDR(c)
		require.NoError(t, err)
		nets = append(nets, n)
	}
	return nets
}

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
	traceRequest(req, mt, cmd, "test-honeypot", "body", nil)

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

	traceRequest(req, mt, parser.Command{}, "login-honeypot", `{"user":"admin"}`, nil)

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

func TestRealClientAddr_NoTrustedProxies_IgnoresHeaders(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.5:4242"
	req.Header.Set("X-Forwarded-For", "8.8.8.8")
	req.Header.Set("X-Real-Ip", "8.8.8.8")

	host, port := realClientAddr(req, nil)

	assert.Equal(t, "10.0.0.5", host)
	assert.Equal(t, "4242", port)
}

func TestRealClientAddr_UntrustedPeer_IgnoresHeaders(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "203.0.113.7:80"
	req.Header.Set("X-Forwarded-For", "8.8.8.8")
	req.Header.Set("X-Real-Ip", "8.8.8.8")

	host, port := realClientAddr(req, mustCIDRs(t, "172.16.0.0/12"))

	assert.Equal(t, "203.0.113.7", host)
	assert.Equal(t, "80", port)
}

func TestRealClientAddr_TrustedPeer_UsesXForwardedFor(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "172.20.0.5:54321"
	req.Header.Set("X-Forwarded-For", "2.39.23.127")

	host, port := realClientAddr(req, mustCIDRs(t, "172.16.0.0/12"))

	assert.Equal(t, "2.39.23.127", host)
	assert.Equal(t, "", port)
}

// XFF poisoning: the attacker prefixes a fake IP, the trusted proxy appends the
// real peer. Walking right-to-left and skipping trusted hops must surface the
// real client (the first non-trusted entry from the right), not the spoof.
func TestRealClientAddr_TrustedPeer_WalksRightToLeftSkippingTrusted(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "172.20.0.5:54321"
	// 1.1.1.1 spoofed by attacker, 2.39.23.127 = real client, 172.20.0.6 = inner trusted hop
	req.Header.Set("X-Forwarded-For", "1.1.1.1, 2.39.23.127, 172.20.0.6")

	host, _ := realClientAddr(req, mustCIDRs(t, "172.16.0.0/12"))

	assert.Equal(t, "2.39.23.127", host)
}

func TestRealClientAddr_TrustedPeer_AllXFFTrusted_FallsBack(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "172.20.0.5:54321"
	req.Header.Set("X-Forwarded-For", "172.20.0.6, 172.20.0.7")

	host, port := realClientAddr(req, mustCIDRs(t, "172.16.0.0/12"))

	assert.Equal(t, "172.20.0.5", host)
	assert.Equal(t, "54321", port)
}

func TestRealClientAddr_TrustedPeer_FallsBackToXRealIp(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "172.20.0.5:54321"
	req.Header.Set("X-Real-Ip", "2.39.23.127")

	host, port := realClientAddr(req, mustCIDRs(t, "172.16.0.0/12"))

	assert.Equal(t, "2.39.23.127", host)
	assert.Equal(t, "", port)
}

func TestRealClientAddr_TrustedPeer_XRealIpIgnoredIfTrusted(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "172.20.0.5:54321"
	req.Header.Set("X-Real-Ip", "172.20.0.9")

	host, port := realClientAddr(req, mustCIDRs(t, "172.16.0.0/12"))

	assert.Equal(t, "172.20.0.5", host)
	assert.Equal(t, "54321", port)
}

func TestRealClientAddr_MalformedXFFEntries_Skipped(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "172.20.0.5:54321"
	req.Header.Set("X-Forwarded-For", "not-an-ip,   ,2.39.23.127, 172.20.0.6")

	host, _ := realClientAddr(req, mustCIDRs(t, "172.16.0.0/12"))

	assert.Equal(t, "2.39.23.127", host)
}

func TestRealClientAddr_IPv6_TrustedPeer(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "[fd00::1]:54321"
	req.Header.Set("X-Forwarded-For", "2001:db8::1234")

	host, _ := realClientAddr(req, mustCIDRs(t, "fd00::/8"))

	assert.Equal(t, "2001:db8::1234", host)
}

func TestRealClientAddr_RemoteAddrWithoutPort(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.5"

	host, port := realClientAddr(req, nil)

	assert.Equal(t, "10.0.0.5", host)
	assert.Equal(t, "", port)
}

func TestTraceRequest_TrustedProxy_ResolvesRealClient(t *testing.T) {
	mt := &mockTracer{}
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.RemoteAddr = "172.20.0.5:54321"
	req.Header.Set("X-Forwarded-For", "2.39.23.127")

	traceRequest(req, mt, parser.Command{Name: "admin"}, "test", "", mustCIDRs(t, "172.16.0.0/12"))

	require.Len(t, mt.events, 1)
	ev := mt.events[0]
	assert.Equal(t, "2.39.23.127", ev.SourceIp)
	assert.Equal(t, "", ev.SourcePort)
	// Raw RemoteAddr is preserved for forensic fidelity.
	assert.Equal(t, "172.20.0.5:54321", ev.RemoteAddr)
}

func TestTraceRequest_UntrustedPeer_DoesNotTrustHeaders(t *testing.T) {
	mt := &mockTracer{}
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.RemoteAddr = "203.0.113.7:8080"
	req.Header.Set("X-Forwarded-For", "8.8.8.8")

	traceRequest(req, mt, parser.Command{}, "test", "", mustCIDRs(t, "172.16.0.0/12"))

	require.Len(t, mt.events, 1)
	assert.Equal(t, "203.0.113.7", mt.events[0].SourceIp)
}
