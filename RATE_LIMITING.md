# Rate Limiting for LLM Plugin

## Overview

IP-based rate limiting for the LLM plugin to prevent API abuse and control costs on public networks.

## Configuration

Rate limiting is **disabled by default** for backward compatibility.

```yaml
plugin:
  llmProvider: "openai"
  llmModel: "gpt-4o"
  openAISecretKey: "sk-proj-xxxxx"
  rateLimitEnabled: true
  rateLimitRequests: 10
  rateLimitWindowSeconds: 60
```

### Parameters

- `rateLimitEnabled` (bool): Enable/disable rate limiting. Default: `false`
- `rateLimitRequests` (int): Max requests per time window
- `rateLimitWindowSeconds` (int): Time window in seconds

## Implementation

- **Algorithm**: Token bucket using `golang.org/x/time/rate`
- **Isolation**: Per-IP rate limiters
- **Thread-safe**: `sync.RWMutex` for concurrent access
- **Response**: Returns "System busy, please try again later" when limit exceeded

## Examples

### Public Research Honeypot
```yaml
rateLimitEnabled: true
rateLimitRequests: 5
rateLimitWindowSeconds: 300  # 5 requests per 5 minutes
```

### Production Honeypot
```yaml
rateLimitEnabled: true
rateLimitRequests: 20
rateLimitWindowSeconds: 60  # 20 requests per minute
```

## Testing

```bash
go test ./plugins/... -v -run TestRateLimit
```

## Logging

Rate limit violations are logged with client IP and command:

```log
WARN Rate limit exceeded client_ip=192.168.1.100 command="ls -la"
```
