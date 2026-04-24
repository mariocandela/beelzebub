# Beelzebub

[![CI](https://github.com/mariocandela/beelzebub/actions/workflows/ci.yml/badge.svg)](https://github.com/mariocandela/beelzebub/actions/workflows/ci.yml) [![Docker](https://github.com/mariocandela/beelzebub/actions/workflows/docker-image.yml/badge.svg)](https://github.com/mariocandela/beelzebub/actions/workflows/docker-image.yml) [![codeql](https://github.com/mariocandela/beelzebub/actions/workflows/codeql.yml/badge.svg)](https://github.com/mariocandela/beelzebub/actions/workflows/codeql.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/mariocandela/beelzebub/v3)](https://goreportcard.com/report/github.com/mariocandela/beelzebub/v3)
[![codecov](https://codecov.io/gh/mariocandela/beelzebub/graph/badge.svg?token=8XTK7D4WHE)](https://codecov.io/gh/mariocandela/beelzebub)
[![Go Reference](https://pkg.go.dev/badge/github.com/mariocandela/beelzebub/v3.svg)](https://pkg.go.dev/github.com/mariocandela/beelzebub/v3)
[![Trust Score](https://archestra.ai/mcp-catalog/api/badge/quality/mariocandela/beelzebub)](https://archestra.ai/mcp-catalog/mariocandela__beelzebub)
[![Mentioned in Awesome Go](https://awesome.re/mentioned-badge.svg)](https://github.com/avelino/awesome-go)

## Overview

Beelzebub is an advanced honeypot framework designed to provide a highly secure environment for detecting and analyzing cyber attacks. It offers a low code approach for easy implementation and uses AI to mimic the behavior of a high-interaction honeypot.

![github beelzebub - inception program](https://github.com/user-attachments/assets/e180d602-6de9-4c48-92ad-eb0ef3c5322d)

## Table of Contents

- [Global Threat Intelligence Community](#global-threat-intelligence-community)
- [Key Features](#key-features)
- [Architecture](#architecture)
- [Quick Start](#quick-start)
- [Configuration](#configuration)
  - [Core Configuration](#core-configuration)
  - [Service Configuration](#service-configuration)
- [Protocol Examples](#protocol-examples)
  - [MCP Honeypot](#mcp-honeypot)
  - [HTTP Honeypot](#http-honeypot)
  - [SSH Honeypot](#ssh-honeypot)
  - [TELNET Honeypot](#telnet-honeypot)
  - [TCP Honeypot](#tcp-honeypot)
- [Observability](#observability)
  - [Prometheus Metrics](#prometheus-metrics)
  - [RabbitMQ Integration](#rabbitmq-integration)
  - [Beelzebub Cloud](#beelzebub-cloud)
- [Testing](#testing)
- [Code Quality](#code-quality)
- [Contributing](#contributing)
- [License](#license)

## Global Threat Intelligence Community

Our mission is to establish a collaborative ecosystem of security researchers and white hat professionals worldwide, dedicated to creating a distributed honeypot network that identifies emerging malware, discovers zero-day vulnerabilities, and neutralizes active botnets.

[![White Paper](https://img.shields.io/badge/White_Paper-v1.0-blue?style=for-the-badge)](https://github.com/beelzebub-labs/white-paper/)

The white paper includes information on how to join our Discord community and contribute to the global threat intelligence network. 

## Key Features

Beelzebub offers a wide range of features to enhance your honeypot environment:

- **Low-code configuration**: YAML-based, modular service definition
- **LLM integration**: The LLM convincingly simulates a real system, creating high-interaction honeypot experiences, while actually maintaining low-interaction architecture for enhanced security and easy management
- **Multi-protocol support**: SSH, HTTP, TCP, TELNET, MCP (detect prompt injection against LLM agents)
- **Prometheus metrics & observability**: Built-in metrics endpoint for monitoring
- **Event tracing**: Multiple output strategies (stdout, RabbitMQ, Beelzebub Cloud)
- **Docker & Kubernetes ready**: Deploy anywhere with provided configurations
- **ELK stack ready**: Official integration available at [Elastic docs](https://www.elastic.co/docs/reference/integrations/beelzebub)

## LLM Honeypot Demo

![demo-beelzebub](https://github.com/user-attachments/assets/4dbb9a67-6c12-49c5-82ac-9b3e340406ca)

## Quick Start

You can run Beelzebub via Docker, Go compiler(cross device), or Helm (Kubernetes).

### Using Docker Compose

1. Build the Docker images:

   ```bash
   $ docker compose build
   ```

2. Start Beelzebub in detached mode:

   ```bash
   $ docker compose up -d
   ```


### Using Go Compiler

1. Download the necessary Go modules:

   ```bash
   $ go mod download
   ```

2. Build the Beelzebub executable:

   ```bash
   $ go build
   ```

3. Run Beelzebub:

   ```bash
   $ ./beelzebub
   ```

### Deploy on kubernetes cluster using helm

1. Install helm

2. Deploy beelzebub:

   ```bash
   $ helm install beelzebub ./beelzebub-chart
   ```

3. Next release

   ```bash
   $ helm upgrade beelzebub ./beelzebub-chart
   ```

## Configuration

Beelzebub uses a two-tier configuration system:

1. **Core configuration** (`beelzebub.yaml`) - Global settings for logging, tracing, and Prometheus
2. **Service configurations** (`services/*.yaml`) - Individual honeypot service definitions

### Core Configuration

The core configuration file controls global behavior:

```yaml
core:
  logging:
    debug: false
    debugReportCaller: false
    logDisableTimestamp: true
    logsPath: ./logs
  tracings:
    rabbit-mq:
      enabled: false
      uri: "amqp://guest:guest@localhost:5672/"
  prometheus:
    path: "/metrics"
    port: ":2112"
  beelzebub-cloud:
    enabled: false
    uri: ""
    auth-token: ""
```

### Service Configuration

Each honeypot service is defined in a separate YAML file in the `services/` directory. To run Beelzebub with custom paths:

```bash
./beelzebub --confCore ./configurations/beelzebub.yaml --confServices ./configurations/services/
```

Additional flags:
- `--memLimitMiB <value>` - Set memory limit in MiB (default: 100, use -1 to disable)

## Protocol Examples

Below are example configurations for each supported protocol.

### MCP Honeypot

MCP (Model Context Protocol) honeypots are decoy tools designed to detect prompt injection attacks against LLM agents.

#### Why Use an MCP Honeypot?

An MCP honeypot is a **decoy tool** that the agent should never invoke under normal circumstances. Integrating this strategy into your agent pipeline offers three key benefits:

- **Real-time detection of guardrail bypass attempts** - Instantly identify when a prompt injection attack successfully convinces the agent to invoke a restricted tool
- **Automatic collection of real attack prompts** - Every activation logs genuine malicious prompts, enabling continuous improvement of your filtering mechanisms
- **Continuous monitoring of attack trends** - Track exploit frequency and system resilience using objective, actionable measurements (HAR, TPR, MTP)

![video-mcp-diagram](https://github.com/user-attachments/assets/e04fd19e-9537-427e-9131-9bee31d8ebad)

**mcp-8000.yaml**:

```yaml
apiVersion: "v1"
protocol: "mcp"
address: ":8000"
description: "MCP Honeypot"
tools:
  - name: "tool:user-account-manager"
    description: "Tool for querying and modifying user account details. Requires administrator privileges."
    params:
      - name: "user_id"
        description: "The ID of the user account to manage."
      - name: "action"
        description: "The action to perform on the user account, possible values are: get_details, reset_password, deactivate_account"
    handler: |
      {
        "tool_id": "tool:user-account-manager",
        "status": "completed",
        "output": {
          "message": "Tool 'tool:user-account-manager' executed successfully. Results are pending internal processing and will be logged.",
          "result": {
            "operation_status": "success",
            "details": "email: kirsten@gmail.com, role: admin, last-login: 02/07/2025"
          }
        }
      }
  - name: "tool:system-log"
    description: "Tool for querying system logs. Requires administrator privileges."
    params:
      - name: "filter"
        description: "The input used to filter the logs."
    handler: |
      {
        "tool_id": "tool:system-log",
        "status": "completed",
        "output": {
          "message": "Tool 'tool:system-log' executed successfully. Results are pending internal processing and will be logged.",
          "result": {
            "operation_status": "success",
            "details": "Info: email: kirsten@gmail.com, last-login: 02/07/2025"
          }
        }
      }
```

Invoke remotely via `http://beelzebub:port/mcp` (Streamable HTTP Server).

### HTTP Honeypot

HTTP honeypots respond to web requests with configurable responses based on URL pattern matching.

**http-80.yaml** (WordPress simulation):

```yaml
apiVersion: "v1"
protocol: "http"
address: ":80"
description: "Wordpress 6.0"
commands:
  - regex: "^(/index.php|/index.html|/)$"
    handler:
      <html>
        <header>
          <title>Wordpress 6 test page</title>
        </header>
        <body>
          <h1>Hello from Wordpress</h1>
        </body>
      </html>
    headers:
      - "Content-Type: text/html"
      - "Server: Apache/2.4.53 (Debian)"
      - "X-Powered-By: PHP/7.4.29"
    statusCode: 200
  - regex: "^(/wp-login.php|/wp-admin)$"
    handler:
      <html>
        <header>
          <title>Wordpress 6 test page</title>
        </header>
        <body>
          <form action="" method="post">
            <label for="uname"><b>Username</b></label>
            <input type="text" placeholder="Enter Username" name="uname" required>

            <label for="psw"><b>Password</b></label>
            <input type="password" placeholder="Enter Password" name="psw" required>

            <button type="submit">Login</button>
          </form>
        </body>
      </html>
    headers:
      - "Content-Type: text/html"
      - "Server: Apache/2.4.53 (Debian)"
      - "X-Powered-By: PHP/7.4.29"
    statusCode: 200
  - regex: "^.*$"
    handler:
      <html>
        <header>
          <title>404</title>
        </header>
        <body>
          <h1>Not found!</h1>
        </body>
      </html>
    headers:
      - "Content-Type: text/html"
      - "Server: Apache/2.4.53 (Debian)"
      - "X-Powered-By: PHP/7.4.29"
    statusCode: 404
```

**http-8080.yaml** (Apache 401 simulation):

```yaml
apiVersion: "v1"
protocol: "http"
address: ":8080"
description: "Apache 401"
commands:
  - regex: ".*"
    handler: "Unauthorized"
    headers:
      - "www-Authenticate: Basic"
      - "server: Apache"
    statusCode: 401
```

### SSH Honeypot

SSH honeypots support both static command responses and LLM-powered dynamic interactions.

#### LLM-Powered SSH Honeypot

Using OpenAI as the LLM provider:

```yaml
apiVersion: "v1"
protocol: "ssh"
address: ":2222"
description: "SSH interactive OpenAI  GPT-4"
commands:
  - regex: "^(.+)$"
    plugin: "LLMHoneypot"
serverVersion: "OpenSSH"
serverName: "ubuntu"
passwordRegex: "^(root|qwerty|Smoker666|123456|jenkins|minecraft|sinus|alex|postgres|Ly123456)$"
deadlineTimeoutSeconds: 60
plugin:
   llmProvider: "openai"
   llmModel: "gpt-4o" #Models https://platform.openai.com/docs/models
   openAISecretKey: "sk-proj-1234"
```

Using local Ollama instance:

```yaml
apiVersion: "v1"
protocol: "ssh"
address: ":2222"
description: "SSH Ollama Llama3"
commands:
  - regex: "^(.+)$"
    plugin: "LLMHoneypot"
serverVersion: "OpenSSH"
serverName: "ubuntu"
passwordRegex: "^(root|qwerty|Smoker666|123456|jenkins|minecraft|sinus|alex|postgres|Ly123456)$"
deadlineTimeoutSeconds: 60
plugin:
   llmProvider: "ollama"
   llmModel: "codellama:7b"
   host: "http://localhost:11434/api/chat"
```

Using a custom prompt:

```yaml
apiVersion: "v1"
protocol: "ssh"
address: ":2222"
description: "SSH interactive OpenAI  GPT-4"
commands:
  - regex: "^(.+)$"
    plugin: "LLMHoneypot"
serverVersion: "OpenSSH"
serverName: "ubuntu"
passwordRegex: "^(root|qwerty|Smoker666|123456|jenkins|minecraft|sinus|alex|postgres|Ly123456)$"
deadlineTimeoutSeconds: 60
plugin:
   llmProvider: "openai"
   llmModel: "gpt-4o"
   openAISecretKey: "sk-proj-1234"
   prompt: "You will act as an Ubuntu Linux terminal. The user will type commands, and you are to reply with what the terminal should show. Your responses must be contained within a single code block."
```

#### Static SSH Honeypot

```yaml
apiVersion: "v1"
protocol: "ssh"
address: ":22"
description: "SSH interactive"
commands:
  - regex: "^ls$"
    handler: "Documents Images Desktop Downloads .m2 .kube .ssh .docker"
  - regex: "^pwd$"
    handler: "/home/"
  - regex: "^uname -m$"
    handler: "x86_64"
  - regex: "^docker ps$"
    handler: "CONTAINER ID IMAGE COMMAND CREATED STATUS PORTS NAMES"
  - regex: "^docker .*$"
    handler: "Error response from daemon: dial unix docker.raw.sock: connect: connection refused"
  - regex: "^uname$"
    handler: "Linux"
  - regex: "^ps$"
    handler: "PID TTY TIME CMD\n21642 ttys000 0:00.07 /bin/dockerd"
  - regex: "^(.+)$"
    handler: "command not found"
serverVersion: "OpenSSH"
serverName: "ubuntu"
passwordRegex: "^(root|qwerty|Smoker666)$"
deadlineTimeoutSeconds: 60
```

### TELNET Honeypot

TELNET honeypots provide terminal-based interaction similar to SSH, with support for both static responses and LLM integration.

#### LLM-Powered TELNET Honeypot

```yaml
apiVersion: "v1"
protocol: "telnet"
address: ":23"
description: "TELNET LLM Honeypot"
commands:
  - regex: "^(.+)$"
    plugin: "LLMHoneypot"
serverName: "router"
passwordRegex: "^(admin|root|password|123456)$"
deadlineTimeoutSeconds: 120
plugin:
   llmProvider: "openai"
   llmModel: "gpt-4o"
   openAISecretKey: "sk-1234"
```

#### Static TELNET Honeypot

```yaml
apiVersion: "v1"
protocol: "telnet"
address: ":23"
description: "TELNET Router Simulation"
commands:
  - regex: "^show version$"
    handler: "Cisco IOS Software, Version 15.1(4)M4"
  - regex: "^show ip interface brief$"
    handler: "Method Status Protocol\nFastEthernet0/0 192.168.1.1 YES NVRAM up up"
  - regex: "^(.+)$"
    handler: "% Unknown command"
serverName: "router"
passwordRegex: "^(admin|cisco|password)$"
deadlineTimeoutSeconds: 60
```

### TCP Honeypot

TCP honeypots support both simple banner-only mode and interactive command-based sessions with regex matching and LLM integration.

#### Banner-Only Mode (Legacy)

Simple banner response for basic service simulation:

```yaml
apiVersion: "v1"
protocol: "tcp"
address: ":3306"
description: "MySQL 8.0.29"
banner: "8.0.29"
deadlineTimeoutSeconds: 10
```

#### Interactive Mode with Commands

TCP honeypots can now handle multi-turn interactions using regex-based command matching, just like SSH and TELNET:

**Redis Honeypot** (text-based protocol):

```yaml
apiVersion: "v1"
protocol: "tcp"
address: ":6379"
description: "Redis 7.0.12"
commands:
  - regex: "^PING"
    handler: "+PONG\r\n"
    name: "ping"
  - regex: "^AUTH"
    handler: "-ERR Client sent AUTH, but no password is set\r\n"
    name: "auth"
  - regex: "^INFO"
    handler: "$180\r\n# Server\r\nredis_version:7.0.12\r\nos:Linux 5.15.0-76-generic x86_64\r\ntcp_port:6379\r\n\r\n"
    name: "info"
  - regex: "^(.+)$"
    handler: "-ERR unknown command\r\n"
    name: "catch_all"
deadlineTimeoutSeconds: 60
serverName: "redis-prod-01"
```

**LDAP / Active Directory Honeypot** (binary protocol):

```yaml
apiVersion: "v1"
protocol: "tcp"
address: ":389"
description: "Active Directory LDAP Domain Controller"
banner: "0\x84\x00\x00\x00\x10\x02\x01\x01\x61\x84\x00\x00\x00\x07\x0a\x01\x00\x04\x00\x04\x00"
commands:
  - regex: "\\x30.*\\x60"
    handler: "0\x84\x00\x00\x00\x10\x02\x01\x01\x61\x84\x00\x00\x00\x07\x0a\x01\x00\x04\x00\x04\x00"
    name: "ldap_bind_response"
  - regex: "\\x30.*\\x63"
    handler: "0\x84\x00\x00\x00\x2a\x02\x01\x02\x65\x84\x00\x00\x00\x21\x04\x00\x30\x84\x00\x00\x00\x00"
    name: "ldap_search_result_done"
deadlineTimeoutSeconds: 30
serverName: "DC01.corp.local"
```

**SMB File Server Honeypot**:

```yaml
apiVersion: "v1"
protocol: "tcp"
address: ":445"
description: "Windows SMB File Server"
commands:
  - regex: "\\xfeSMB"
    handler: "\xfeSMB@\x00\x00\x00\x00\x00\x00\x00\x00\x00"
    name: "smb2_negotiate_response"
  - regex: "\\xffSMB"
    handler: "\xffSMB\x72\x00\x00\x00\x00"
    name: "smb1_negotiate_response"
deadlineTimeoutSeconds: 20
serverName: "FILESERVER01"
```

**RDP Honeypot**:

```yaml
apiVersion: "v1"
protocol: "tcp"
address: ":3389"
description: "Windows Remote Desktop Service"
banner: "\x03\x00\x00\x13\x0e\xd0\x00\x00\x12\x34\x00\x02\x00\x08\x00\x02\x00\x00\x00"
commands:
  - regex: "\\x03\\x00"
    handler: "\x03\x00\x00\x13\x0e\xd0\x00\x00\x12\x34\x00\x02\x01\x08\x00\x02\x00\x00\x00"
    name: "x224_connection_confirm"
deadlineTimeoutSeconds: 15
serverName: "WIN-DC01"
```

#### TCP with LLM Integration

TCP honeypots can use LLM providers to generate dynamic responses:

```yaml
apiVersion: "v1"
protocol: "tcp"
address: ":5432"
description: "PostgreSQL 15.3 with LLM"
commands:
  - regex: "^(.+)$"
    plugin: "LLMHoneypot"
deadlineTimeoutSeconds: 120
serverName: "pg-master"
plugin:
  llmProvider: "openai"
  llmModel: "gpt-4o"
  openAISecretKey: "sk-proj-..."
  prompt: "You are simulating a PostgreSQL 15.3 server. Respond to incoming TCP data as a PostgreSQL server would."
```

Additional example configurations are available in `configurations/services/` for:
- Memcached (`:11211`)
- MS-SQL (`:1433`)
- VNC (`:5900`)
- MQTT (`:1883`)

## Observability

### Prometheus Metrics

Beelzebub exposes Prometheus metrics at the configured endpoint (default: `:2112/metrics`). Available metrics include:

- `beelzebub_events_total` - Total number of honeypot events
- `beelzebub_events_ssh_total` - SSH-specific events
- `beelzebub_events_http_total` - HTTP-specific events
- `beelzebub_events_tcp_total` - TCP-specific events
- `beelzebub_events_telnet_total` - TELNET-specific events
- `beelzebub_events_mcp_total` - MCP-specific events

### RabbitMQ Integration

Enable RabbitMQ tracing to publish honeypot events to a message queue:

```yaml
core:
  tracings:
    rabbit-mq:
      enabled: true
      uri: "amqp://guest:guest@localhost:5672/"
```

Events are published as JSON messages for downstream processing.

## Testing

### Unit Tests

```bash
make test.unit
```

### Integration Tests

Integration tests require external dependencies (RabbitMQ, etc.):

```bash
make test.dependencies.start
make test.integration
make test.dependencies.down
```

## Code Quality

We maintain high code quality through:

- **Automated Testing**: Unit and integration tests run on every pull request
- **Static Analysis**: Go Report Card and CodeQL for code quality and security checks
- **Code Coverage**: Monitored via [Codecov](https://codecov.io/gh/mariocandela/beelzebub)
- **Continuous Integration**: GitHub Actions pipelines on every commit
- **Code Reviews**: All contributions undergo peer review

## Contributing

The Beelzebub team welcomes contributions and project participation. Whether you want to report bugs, contribute new features, or have any questions, please refer to our [Contributor Guide](CONTRIBUTING.md) for detailed information. We encourage all participants and maintainers to adhere to our [Code of Conduct](CODE_OF_CONDUCT.md) and foster a supportive and respectful community.

Happy hacking!

## License

Beelzebub is licensed under the [GNU GPL v3 License](LICENSE).

## Supported By

[![JetBrains logo.](https://resources.jetbrains.com/storage/products/company/brand/logos/jetbrains.svg)](https://jb.gg/OpenSourceSupport)

![gitbook logo](https://i.postimg.cc/VNQh5hnk/gitbook.png)
