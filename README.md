# Beelzebub

[![CI](https://github.com/mariocandela/beelzebub/actions/workflows/ci.yml/badge.svg)](https://github.com/mariocandela/beelzebub/actions/workflows/ci.yml) [![Docker](https://github.com/mariocandela/beelzebub/actions/workflows/docker-image.yml/badge.svg)](https://github.com/mariocandela/beelzebub/actions/workflows/docker-image.yml) [![codeql](https://github.com/mariocandela/beelzebub/actions/workflows/codeql.yml/badge.svg)](https://github.com/mariocandela/beelzebub/actions/workflows/codeql.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/mariocandela/beelzebub/v3)](https://goreportcard.com/report/github.com/mariocandela/beelzebub/v3)
[![codecov](https://codecov.io/gh/mariocandela/beelzebub/graph/badge.svg?token=8XTK7D4WHE)](https://codecov.io/gh/mariocandela/beelzebub)
[![Go Reference](https://pkg.go.dev/badge/github.com/mariocandela/beelzebub/v3.svg)](https://pkg.go.dev/github.com/mariocandela/beelzebub/v3)
[![Mentioned in Awesome Go](https://awesome.re/mentioned-badge.svg)](https://github.com/avelino/awesome-go)  

## Overview

Beelzebub is an advanced honeypot framework designed to provide a highly secure environment for detecting and analyzing cyber attacks. It offers a low code approach for easy implementation and uses AI to mimic the behavior of a high-interaction honeypot.

<img src="https://beelzebub.netlify.app/go-beelzebub.png" alt="Beelzebub Logo" width="200"/>

## LLM Honeypot

[![asciicast](https://asciinema.org/a/665295.svg)](https://asciinema.org/a/665295)


## Telegram Bot for Real-Time Attacks

Stay updated on real-time attacks by joining our dedicated Telegram channel: [Telegram Channel](https://t.me/beelzebubhoneypot)

## Examples

To better understand the capabilities of Beelzebub, you can explore our example repository: [mariocandela/beelzebub-example](https://github.com/mariocandela/beelzebub-example)

## Quick Start

We provide two quick start options for build and run Beelzebub: using Docker Compose or the Go compiler.

### Using Docker Compose

1. Build the Docker images:

   ```bash
   $ docker-compose build
   ```

2. Start Beelzebub in detached mode:

   ```bash
   $ docker-compose up -d
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
## Testing

We provide two types of tests: unit tests and integration tests.

### Unit Tests

To run unit tests:

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

## Key Features

Beelzebub offers a wide range of features to enhance your honeypot environment:

- Support for Ollama
- Support for OpenAI
- SSH Honeypot
- HTTP Honeypot
- TCP Honeypot
- Prometheus openmetrics integration
- Docker integration
- RabbitMQ integration
- kubernetes

## Example Configuration

Beelzebub allows easy configuration for different services and ports. Simply create a new file for each service/port within the `/configurations/services` directory.

To execute Beelzebub with your custom path, use the following command:

```bash
$ ./beelzebub --confCore ./configurations/beelzebub.yaml --confServices ./configurations/services/
```

Here are some example configurations for different honeypot scenarios:

#### Example HTTP Honeypot on Port 80

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

#### Example HTTP Honeypot on Port 8080

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

#### Example SSH Honeypot

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

###### SSH Honeypot on Port 22

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

## Roadmap

Our future plans for Beelzebub include developing it into a robust PaaS platform.

## Contributing

The Beelzebub team welcomes contributions and project participation. Whether you want to report bugs, contribute new features, or have any questions, please refer to our [Contributor Guide](CONTRIBUTING.md) for detailed information. We encourage all participants and maintainers to adhere to our [Code of Conduct](CODE_OF_CONDUCT.md) and foster a supportive and respectful community.

Happy hacking!

## License

Beelzebub is licensed under the [MIT License](LICENSE).

## Supported by
[![JetBrains logo.](https://resources.jetbrains.com/storage/products/company/brand/logos/jetbrains.svg)](https://jb.gg/OpenSourceSupport)

![gitbook logo](https://i.postimg.cc/VNQh5hnk/gitbook.png)