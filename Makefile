include .env

# ==================================================================================== #
# HELPERS
# ==================================================================================== #

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]

# ==================================================================================== #
# DEVELOPMENT
# ==================================================================================== #


## db/psql: connect to the database using psql
.PHONY: db/psql
db/psql:
	psql ${HELLO_DB_DSN}

## db/tables: list tables in the connected database
.PHONY: db/tables
db/tables:
	psql ${HELLO_DB_DSN} -c "\dt"

## db/migrations/new name=$1: create a new database migration
.PHONY: db/migrations/new
db/migrations/new:
	@echo 'Creating migration files for ${name}...'
	migrate create -seq -ext=.sql -dir=./migrations ${name}

## test/setup: copy migrations
.PHONY: test/setup
test/setup:
	@cp -r migrations/ cmd/api/migrations/

## test/cleanup: remove migrations
.PHONY: test/cleanup
test/cleanup:
	@rm -rf cmd/api/migrations

## test: run all tests with verbose output
.PHONY: test
test: test/setup
	@go test -v ./...
	@$(MAKE) test/cleanup

## test/short: run tests excluding slow ones with verbose output
.PHONY: test/short
test/short: test/setup
	@go test -v -short ./...
	@$(MAKE) test/cleanup

## compose/up: start docker compose services
.PHONY: compose/up
compose/up:
	docker compose up --build -d

## compose/down: stop docker compose services and remove volumes
.PHONY: compose/down
compose/down:
	docker compose down -v

## logs: show application logs from docker container
.PHONY: logs
logs:
	docker compose logs -f app

## setup/dev: complete development environment setup
.PHONY: setup/dev
setup/dev: compose/up
	@echo "Waiting for PostgreSQL to be ready...\n"
	@sleep 3
	@echo "📋 Listing database tables...\n"
	@$(MAKE) db/tables
	@echo '✅ Development environment ready!'
