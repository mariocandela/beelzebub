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

test.dependencies.start:
	${DOCKER_COMPOSE} -f ./integration_test/docker-compose.yml up -d

test.dependencies.down:
	${DOCKER_COMPOSE} -f ./integration_test/docker-compose.yml down

test.integration:
	INTEGRATION=1 go test ./...

test.integration.verbose:
	INTEGRATION=1 go test ./... -v

# .PHONY : is an idiomatic way to differentiate commands from files in GNU Make
.PHONY:
	
	beelzebub.start

.PHONY:
	
	beelzebub.stop

.PHONY:

	test.unit

.PHONY:

	test.dependencies.start

.PHONY:

	test.dependencies.down

.PHONY:

	test.integration

.PHONY:

	test.integration.verbose
