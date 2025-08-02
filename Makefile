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

## run/api: run the cmd/api application
.PHONY: run/api
run/api:
	go run ./cmd/api -db-dsn=${HELLO_DB_DSN}

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

## db/migrations/up: apply all up database migrations
.PHONY: db/migrations/up
db/migrations/up: confirm
	@echo 'Running up migrations...'
	migrate -path ./migrations -database ${HELLO_DB_DSN} up

## db/migrations/down: apply all down database migrations
.PHONY: db/migrations/down
db/migrations/down: confirm
	@echo 'Running down migrations...'
	migrate -path ./migrations -database ${HELLO_DB_DSN} down

## compose/up: start docker compose services
.PHONY: compose/up
compose/up:
	docker compose up -d

## compose/down: stop docker compose services and remove volumes
.PHONY: compose/down
compose/down:
	docker compose down -v


## setup/dev: complete development environment setup
.PHONY: setup/dev
setup/dev: compose/up
	@echo "Waiting for PostgreSQL to be ready...\n"
	@sleep 3
	@echo "💡 Applying database migrations...\n"
	@$(MAKE) db/migrations/up
	@echo "📋 Listing database tables...\n"
	@$(MAKE) db/tables
	@echo '✅ Development environment ready! Run "make run/api" to start the server.'
