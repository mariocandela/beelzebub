DOCKER_COMPOSE := $(shell which docker-compose)

ifeq (${DOCKER_COMPOSE},)
DOCKER_COMPOSE = docker compose
endif

beelzebub.start:
	${DOCKER_COMPOSE} build;
	${DOCKER_COMPOSE} up -d;

beelzebub.stop:
	${DOCKER_COMPOSE} down;

test.unit:
	go test ./...

test.unit.verbose:
	go test ./... -v

test.integration:
	INTEGRATION=1 go test ./...

test.integration.verbose:
	INTEGRATION=1 go test ./... -v