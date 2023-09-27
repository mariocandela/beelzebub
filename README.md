# Beelzebub

[![CI](https://github.com/mariocandela/beelzebub/actions/workflows/ci.yml/badge.svg)](https://github.com/mariocandela/beelzebub/actions/workflows/ci.yml) [![Docker](https://github.com/mariocandela/beelzebub/actions/workflows/docker-image.yml/badge.svg)](https://github.com/mariocandela/beelzebub/actions/workflows/docker-image.yml) [![codeql](https://github.com/mariocandela/beelzebub/actions/workflows/codeql.yml/badge.svg)](https://github.com/mariocandela/beelzebub/actions/workflows/codeql.yml)

## Overview

Beelzebub is an advanced honeypot framework designed to provide a highly secure environment for detecting and analyzing cyber attacks. It offers a low code approach for easy implementation and utilizes virtualization techniques powered by OpenAI Generative Pre-trained Transformer.

<img src="https://beelzebub.netlify.app/go-beelzebub.png" alt="Beelzebub Logo" width="200"/>

## OpenAI GPT Integration

Learn how to integrate Beelzebub with OpenAI GPT-3 by referring to our comprehensive guide on Medium: [Medium Article](https://medium.com/@mario.candela.personal/how-to-build-a-highly-effective-honeypot-with-beelzebub-and-chatgpt-a2f0f05b3e1)

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
$ make test.integration
```

## Key Features

Beelzebub offers a wide range of features to enhance your honeypot environment:

- OpenAI Generative Pre-trained Transformer act as Linux virtualization
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

###### Honeypot with GPT-3 on Port 2222

```yaml
apiVersion: "v1"
protocol: "ssh"
address: ":2222"
description: "SSH interactive ChatGPT"
commands:
  - regex: "^(.+)$"
    plugin: "OpenAIGPTLinuxTerminal"
serverVersion: "OpenSSH"
serverName: "ubuntu"
passwordRegex: "^(root|qwerty|Smoker666|123456|jenkins|minecraft|sinus|alex|postgres|Ly123456)$"
deadlineTimeoutSeconds: 60
plugin:
  openAPIChatGPTSecretKey: "Your OpenAI Secret Key"
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

[![asciicast](https://asciinema.org/a/604522.svg)](https://asciinema.org/a/604522)

## Roadmap

Our future plans for Beelzebub include developing it into a robust PaaS platform.

## Contributing

The Beelzebub team welcomes contributions and project participation. Whether you want to report bugs, contribute new features, or have any questions, please refer to our [Contributor Guide](CONTRIBUTING.md) for detailed information. We encourage all participants and maintainers to adhere to our [Code of Conduct](CODE_OF_CONDUCT.md) and foster a supportive and respectful community.

Happy hacking!

## License

Beelzebub is licensed under the [MIT License](LICENSE).

## Supported by JetBrains
[![JetBrains Black Box Logo logo](https://resources.jetbrains.com/storage/products/company/brand/logos/jb_square.png)](https://jb.gg/OpenSourceSupport)
