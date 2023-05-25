DOCKER_COMPOSE := $(shell which docker-compose)

ifeq (${DOCKER_COMPOSE},)
DOCKER_COMPOSE = docker compose
endif

.PHONY: beelzebub.start
beelzebub.start:
	${DOCKER_COMPOSE} build;
	${DOCKER_COMPOSE} up -d;

.PHONY: beelzebub.stop
beelzebub.stop:
	${DOCKER_COMPOSE} down;

.PHONY: test.unit
test.unit:
	go test ./...

.PHONY: test.unit.verbose
test.unit.verbose:
	go test ./... -v

.PHONY: test.dependencies.start
test.dependencies.start:
	${DOCKER_COMPOSE} -f ./integration_test/docker-compose.yml up -d

.PHONY:	test.dependencies.down
test.dependencies.down:
	${DOCKER_COMPOSE} -f ./integration_test/docker-compose.yml down

.PHONY: test.integration
test.integration:
	INTEGRATION=1 go test ./...

.PHONY: test.integration.verbose
test.integration.verbose:
	INTEGRATION=1 go test ./... -v
