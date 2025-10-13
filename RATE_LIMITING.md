# Rate Limiting for LLM Plugin

## Overview

The LLM plugin now supports IP-based rate limiting to prevent abuse and control costs when the honeypot is exposed on public networks. This feature uses a token bucket algorithm to limit the number of LLM API calls per IP address.

## Configuration

Rate limiting is configured in the service YAML files under the `plugin` section. It is **disabled by default** to maintain backward compatibility.

### Configuration Parameters

- `rateLimitEnabled` (bool): Enable or disable rate limiting. Default: `false`
- `rateLimitRequests` (int): Maximum number of requests allowed per time window
- `rateLimitWindowSeconds` (int): Time window in seconds for the rate limit

### Example Configuration

```yaml
apiVersion: "v1"
protocol: "ssh"
address: ":2222"
description: "SSH interactive ChatGPT with rate limiting"
commands:
  - regex: "^(.+)$"
    plugin: "LLMHoneypot"
serverVersion: "OpenSSH"
serverName: "ubuntu"
passwordRegex: "^(root|qwerty|Smoker666)$"
deadlineTimeoutSeconds: 60
plugin:
  llmProvider: "openai"
  llmModel: "gpt-4o"
  openAISecretKey: "sk-proj-xxxxx"
  rateLimitEnabled: true
  rateLimitRequests: 10
  rateLimitWindowSeconds: 60
```

This configuration allows 10 requests per 60 seconds per IP address.

## How It Works

1. **Per-IP Tracking**: Each client IP address gets its own rate limiter instance
2. **Token Bucket Algorithm**: Uses `golang.org/x/time/rate` for smooth rate limiting
3. **Automatic Response**: When rate limit is exceeded, returns "System busy, please try again later" instead of making an LLM API call
4. **Thread-Safe**: Uses `sync.RWMutex` for concurrent access to rate limiters

## Behavior

- When rate limiting is **disabled**: All requests are processed normally
- When rate limiting is **enabled**:
  - Requests within the limit are processed normally
  - Requests exceeding the limit receive a generic "busy" message
  - Each IP address has its own independent rate limit
  - Rate limiters are created on-demand for new IP addresses

## Logging

When a rate limit is exceeded, a warning is logged with the following fields:

- `client_ip`: The IP address that exceeded the limit
- `command`: The command that was attempted

Example log entry:

```text
WARN Rate limit exceeded client_ip=192.168.1.100 command="ls -la"
```

## Use Cases

### Research Honeypot on Public Network

```yaml
rateLimitEnabled: true
rateLimitRequests: 5
rateLimitWindowSeconds: 300 # 5 requests per 5 minutes
```

### Development/Testing Environment

```yaml
rateLimitEnabled: false # No rate limiting for testing
```

### High-Traffic Production Honeypot

```yaml
rateLimitEnabled: true
rateLimitRequests: 20
rateLimitWindowSeconds: 60 # 20 requests per minute
```

## Implementation Details

- Rate limiters are stored in a map keyed by client IP address
- The token bucket refills at a constant rate: `rateLimitRequests / rateLimitWindowSeconds`
- Burst capacity equals `rateLimitRequests` to allow initial bursts
- Memory usage scales with the number of unique IP addresses that connect

## Testing

Run the rate limiting tests:

```bash
go test ./plugins/... -v -run TestRateLimit
```

## Backward Compatibility

This feature is fully backward compatible:

- Existing configurations without rate limiting parameters will work unchanged
- Rate limiting is disabled by default
- No breaking changes to the API or configuration format
