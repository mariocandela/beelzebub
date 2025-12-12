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

## 🌍 Global Threat Intelligence Community

Our mission is to establish a collaborative ecosystem of security researchers and white hat professionals worldwide, dedicated to creating a distributed honeypot network that identifies emerging malware, discovers zero-day vulnerabilities, and neutralizes active botnets. 

For a comprehensive overview of our distributed threat intelligence framework and community vision, please refer to our white paper:

[![White Paper](https://img.shields.io/badge/White_Paper-v1.0-blue?style=for-the-badge)](https://github.com/beelzebub-labs/white-paper/)

*The white paper includes information on how to join our Discord community and contribute to the global threat intelligence network.* 

## Key Features

Beelzebub offers a wide range of features to enhance your honeypot environment:

- Low-code configuration: YAML-based, modular service definition
- LLM integration: The LLM convincingly simulates a real system, creating high-interaction honeypot experiences, while actually maintaining low-interaction architecture for enhanced security and easy management.
- Multi-protocol support: SSH, HTTP, TCP, MCP(Detect prompt injection against LLM agents)
- Prometheus metrics & observability 
- Docker & Kubernetes ready
- ELK stack ready, docs: [Official ELK integration](https://www.elastic.co/docs/reference/integrations/beelzebub)

## LLM Honeypot Demo
![demo-beelzebub](https://github.com/user-attachments/assets/4dbb9a67-6c12-49c5-82ac-9b3e340406ca)

## Code Quality

We are strongly committed to maintaining high code quality in the Beelzebub project. Our development workflow includes comprehensive testing, code reviews, static analysis, and continuous integration to ensure the reliability and maintainability of the codebase.

### What We Do

* **Automated Testing:**
  Both unit and integration tests are run on every pull request to catch regressions and ensure stability.

* **Static Analysis:**
  We use tools like Go Report Card and CodeQL to automatically check for code quality, style, and security issues.

* **Code Coverage:**
  Our test coverage is monitored with [Codecov](https://codecov.io/gh/mariocandela/beelzebub), and we aim for extensive coverage of all core components.

* **Continuous Integration:**
  Every commit triggers automated CI pipelines on GitHub Actions, which run all tests and quality checks.

* **Code Reviews:**
  All new contributions undergo peer review to maintain consistency and high standards across the project.

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

## Example Configuration

Beelzebub allows easy configuration for different services and ports. Simply create a new file for each service/port within the `/configurations/services` directory.

To execute Beelzebub with your custom path, use the following command:

```bash
$ ./beelzebub --confCore ./configurations/beelzebub.yaml --confServices ./configurations/services/
```

Here are some example configurations for different honeypot scenarios:

### MCP Honeypot

#### Why choose an MCP Honeypot?

An MCP honeypot is a **decoy tool** that the agent should never invoke under normal circumstances. Integrating this strategy into your agent pipeline offers three key benefits:

* **Real-time detection of guardrail bypass attempts.**
  
  Instantly identify when a prompt injection attack successfully convinces the agent to invoke a restricted tool.
* **Automatic collection of real attack prompts for guardrail fine-tuning.**
  
   Every activation logs genuine malicious prompts, enabling continuous improvement of your filtering mechanisms.
* **Continuous monitoring of attack trends through key metrics (HAR, TPR, MTP).**
  
   Track exploit frequency and system resilience using objective, actionable measurements.

![video-mcp-diagram](https://github.com/user-attachments/assets/e04fd19e-9537-427e-9131-9bee31d8ebad)

##### Example MCP Honeypot Configuration

###### mcp-8000.yaml

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

#### Invoke remotely: beelzebub:port/mcp (Streamable HTTPServer).

### HTTP Honeypot

###### http-80.yaml

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

### HTTP Honeypot

###### http-8080.yaml

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

###### LLM Honeypots

Follow a SSH LLM Honeypot using OpenAI as provider LLM:

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
   openAISecretKey: "sk-proj-123456"
```

Examples with local Ollama instance using model codellama:7b:

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
   llmModel: "codellama:7b" #Models https://ollama.com/search
   host: "http://example.com/api/chat" #default http://localhost:11434/api/chat
```
Example with custom prompt:

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
   openAISecretKey: "sk-proj-123456"
   prompt: "You will act as an Ubuntu Linux terminal. The user will type commands, and you are to reply with what the terminal should show. Your responses must be contained within a single code block."
```

###### SSH Honeypot

###### ssh-22.yaml

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

## Testing

Maintaining excellent code quality is essential for security-focused projects like Beelzebub. We welcome all contributors who share our commitment to robust, readable, and reliable code!

### Unit Tests

For contributor, we have a comprehensive suite of unit/integration tests that cover the core functionality of Beelzebub. To run the unit tests, use the following command:

```bash
$ make test.unit
```

### Integration Tests

To run integration tests:

```bash
$ make test.dependencies.start
$ make test.integration
$ make test.dependencies.down
```

## Contributing

The Beelzebub team welcomes contributions and project participation. Whether you want to report bugs, contribute new features, or have any questions, please refer to our [Contributor Guide](CONTRIBUTING.md) for detailed information. We encourage all participants and maintainers to adhere to our [Code of Conduct](CODE_OF_CONDUCT.md) and foster a supportive and respectful community.

Happy hacking!

## License

Beelzebub is licensed under the [GNU GPL v3 License](LICENSE).

## Supported by
[![JetBrains logo.](https://resources.jetbrains.com/storage/products/company/brand/logos/jetbrains.svg)](https://jb.gg/OpenSourceSupport)

![gitbook logo](https://i.postimg.cc/VNQh5hnk/gitbook.png)
