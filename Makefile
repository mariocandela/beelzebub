DOCKER_COMPOSE := $(shell which docker-compose)

ifeq (${DOCKER_COMPOSE},)
DOCKER_COMPOSE = docker compose
endif

INTEGRATION_TEST_PATH?=./integration_test

docker.start.components:
	${DOCKER_COMPOSE} up -d rabbitmq

docker.stop:
	${DOCKER_COMPOSE} down;

test.integration:
	go test -tags=integration $(INTEGRATION_TEST_PATH) -count=1 -run

test.integration.verbose:
	go test -tags=integration $(INTEGRATION_TEST_PATH) -count=1 -v -run