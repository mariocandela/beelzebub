![CI](https://github.com/mariocandela/beelzebub/actions/workflows/ci.yml/badge.svg) ![Docker](https://github.com/mariocandela/beelzebub/actions/workflows/docker-image.yml/badge.svg) ![codeql](https://github.com/mariocandela/beelzebub/actions/workflows/codeql.yml/badge.svg)
# Beelzebub
[![logo-1.png](https://i.postimg.cc/KvbsJFp3/logo-1.png)](https://postimg.cc/yWfPNqH7)

A secure honeypot framework low code, extremely easy to configure by yaml 🚀

## OpenAI GPT integration
How to integrate with OpenAI GPT-3: [`Medium Article`](https://medium.com/@mario.candela.personal/how-to-build-a-highly-effective-honeypot-with-beelzebub-and-chatgpt-a2f0f05b3e1)

[![OpenAI Integration Diagram](https://static.swimlanes.io/24d6634a381aa8eb0decf5bac7ae214d.png)](https://static.swimlanes.io/24d6634a381aa8eb0decf5bac7ae214d.png)

## Telegram bot realtime attacks

 bot: [`telegram channel`](https://t.me/beelzebubhoneypot)

## Examples 

[`mariocandela/beelzebub-example`](https://github.com/mariocandela/beelzebub-example)

## Quick Start

Using [`docker-compose`](https://docs.docker.com/compose/)

```bash
$ docker-compose build
$ docker-compose up -d
 ```

Using [`go compiler`](https://go.dev/doc/install)

```bash
$ go mod download
$ go build 
$ ./beelzebub
 ```

### Unit Test:

```bash
$ make test.unit
 ```

### Integration test:

Run integration testing
```bash
$ make test.integration
 ```


## Features

- OpenAPI ChatBot GPT-3 Linux virtualization 
- SSH Honeypot
- HTTP Honeypot
- TCP Honeypot
- Prometheus openmetrics
- Docker
- RabbitMQ integration

## Example configuration service 

The configurations are inside the /configurations/services directory, just add a new file for each service/port.

### Example HTTP Honeypot on 80 port

###### http-80.yaml

```yaml
apiVersion: "v1"
protocol: "http"
address: ":80"
description: "Wordpress 6.0"
commands:
  - regex: "index.php"
    handler: ""
    headers:
      - "Content-Type: text/html"
      - "Server: Apache/2.4.53 (Debian)"
      - "X-Powered-By: PHP/7.4.29"
    statusCode: 200
  - regex: "^(wp-login.php|/wp-admin)$"
    handler: ""
    headers:
      - "Content-Type: text/html"
      - "Server: Apache/2.4.53 (Debian)"
      - "X-Powered-By: PHP/7.4.29"
    statusCode: 200
 ```

![alt text](https://i.postimg.cc/529V6jYz/Schermata-2022-06-02-alle-12-42-46.png)


### Example HTTP Honeypot on 8080 port

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

![alt text](https://i.postimg.cc/T1cs6qc4/Schermata-2022-06-02-alle-12-43-55.png)

### Example SSH Honeypot

###### Honeypot with ChatBot GPT-3 ssh-2222.yaml

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
  openAPIChatGPTSecretKey: "Here your ChatBot SecretKey "
 ```

###### ssh-22.yaml

```yaml
apiVersion: "v1"
protocol: "ssh"
address: ":22"
description: "SSH interactive"
commands:
  - regex: "^ls$"
    handler: "Documents Images  Desktop Downloads .m2 .kube .ssh  .docker"
  - regex: "^pwd$"
    handler: "/home/"
  - regex: "^uname -m$"
    handler: "x86_64"
  - regex: "^docker ps$"
    handler: "CONTAINER ID   IMAGE     COMMAND   CREATED   STATUS    PORTS     NAMES"
  - regex: "^docker .*$"
    handler: "Error response from daemon: dial unix docker.raw.sock: connect: connection refused"
  - regex: "^uname$"
    handler: "Linux"
  - regex: "^ps$"
    handler: "  PID TTY           TIME CMD\n21642 ttys000    0:00.07 /bin/dockerd"
  - regex: "^(.+)$"
    handler: "command not found"
serverVersion: "OpenSSH"
serverName: "ubuntu"
passwordRegex: "^(root|qwerty|Smoker666)$"
deadlineTimeoutSeconds: 60
 ```

![alt text](https://i.postimg.cc/jdpfT0LB/Schermata-2022-06-02-alle-12-46-50.png)

## TODO

- telnet
- UDP

# ROADMAP

- SaaS Platform


## Documentation

- [API Docs](https://) #TODO

## Contributing

The beelzebub team enthusiastically welcomes contributions and project participation! There's a bunch of things you can do if you want to contribute! The [Contributor Guide](CONTRIBUTING.md) has all the information you need for everything from reporting bugs to contributing entire new features. Please don't hesitate to jump in if you'd like to, or even ask us questions if something isn't clear.

All participants and maintainers in this project are expected to follow [Code of Conduct](CODE_OF_CONDUCT.md), and just generally be excellent to each other.

Happy hacking!

## License

This project is licensed under [GNU GPL 3 License](LICENSE).

[![](https://www.paypalobjects.com/en_US/i/btn/btn_donateCC_LG.gif)](https://www.paypal.com/donate/?business=P75FH5LXKQTAC&no_recurring=0&currency_code=EUR)
