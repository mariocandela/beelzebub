package HTTP

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseLLMHTTPResponseValidJSON(t *testing.T) {
	resp := &httpResponse{StatusCode: 200}
	completions := `{"headers":{"Content-Type":"text/html","Server":"Apache/2.4.41","X-Custom":"value"},"body":"<html><body>Hello</body></html>","statusCode":200}`

	parseLLMHTTPResponse(completions, resp)

	assert.Equal(t, "<html><body>Hello</body></html>", resp.Body)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Contains(t, resp.Headers, "Content-Type: text/html")
	assert.Contains(t, resp.Headers, "Server: Apache/2.4.41")
	assert.Contains(t, resp.Headers, "X-Custom: value")
}

func TestParseLLMHTTPResponseWithStatusCode(t *testing.T) {
	resp := &httpResponse{StatusCode: 200}
	completions := `{"headers":{"Content-Type":"application/json"},"body":"{\"error\":\"not found\"}","statusCode":404}`

	parseLLMHTTPResponse(completions, resp)

	assert.Equal(t, `{"error":"not found"}`, resp.Body)
	assert.Equal(t, 404, resp.StatusCode)
	assert.Contains(t, resp.Headers, "Content-Type: application/json")
}

func TestParseLLMHTTPResponseFallbackPlainText(t *testing.T) {
	resp := &httpResponse{StatusCode: 200}
	completions := "just plain text response from the LLM"

	parseLLMHTTPResponse(completions, resp)

	assert.Equal(t, "just plain text response from the LLM", resp.Body)
	assert.Empty(t, resp.Headers)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestParseLLMHTTPResponseFallbackInvalidJSON(t *testing.T) {
	resp := &httpResponse{StatusCode: 200}
	completions := `{"headers": "not a map", broken json`

	parseLLMHTTPResponse(completions, resp)

	assert.Equal(t, completions, resp.Body)
	assert.Empty(t, resp.Headers)
}

func TestParseLLMHTTPResponseFallbackHTMLDirectly(t *testing.T) {
	resp := &httpResponse{StatusCode: 200}
	completions := "<html><body><h1>404 Not Found</h1></body></html>"

	parseLLMHTTPResponse(completions, resp)

	assert.Equal(t, completions, resp.Body)
	assert.Empty(t, resp.Headers)
}

func TestParseLLMHTTPResponseEmptyHeaders(t *testing.T) {
	resp := &httpResponse{StatusCode: 200}
	completions := `{"headers":{},"body":"empty headers response"}`

	parseLLMHTTPResponse(completions, resp)

	assert.Equal(t, "empty headers response", resp.Body)
	assert.Empty(t, resp.Headers)
}

func TestParseLLMHTTPResponseNoStatusCodeKeepsDefault(t *testing.T) {
	resp := &httpResponse{StatusCode: 201}
	completions := `{"headers":{"Server":"nginx"},"body":"created"}`

	parseLLMHTTPResponse(completions, resp)

	assert.Equal(t, "created", resp.Body)
	assert.Equal(t, 201, resp.StatusCode)
	assert.Contains(t, resp.Headers, "Server: nginx")
}

func TestParseLLMHTTPResponseSkipsManagedHeaders(t *testing.T) {
	resp := &httpResponse{StatusCode: 200}
	completions := `{"headers":{"Content-Type":"text/html","Content-Length":"999","Date":"wrong","Transfer-Encoding":"chunked","Connection":"close","Server":"nginx"},"body":"test"}`

	parseLLMHTTPResponse(completions, resp)

	assert.Equal(t, "test", resp.Body)
	assert.Contains(t, resp.Headers, "Content-Type: text/html")
	assert.Contains(t, resp.Headers, "Server: nginx")
	for _, h := range resp.Headers {
		lower := strings.ToLower(h)
		assert.NotContains(t, lower, "content-length")
		assert.NotContains(t, lower, "date:")
		assert.NotContains(t, lower, "transfer-encoding")
		assert.NotContains(t, lower, "connection:")
	}
}

func TestParseLLMHTTPResponsePreservesExistingHeaders(t *testing.T) {
	resp := &httpResponse{
		StatusCode: 200,
		Headers:    []string{"X-Existing: keep-me"},
	}
	completions := `{"headers":{"X-New":"added"},"body":"test"}`

	parseLLMHTTPResponse(completions, resp)

	assert.Equal(t, "test", resp.Body)
	assert.Contains(t, resp.Headers, "X-Existing: keep-me")
	assert.Contains(t, resp.Headers, "X-New: added")
}
