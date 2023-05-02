include .env

# ==================================================================================== # 
# HELPERS
# ==================================================================================== #

## help: print this help messages
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

.PHONY: confirm
confirm:
	@echo 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]


# ==================================================================================== # 
# DEVELOPMENT
# ==================================================================================== #

## run/api: run the cmd/api application
.PHONY: run/api
run/api:
	go run ./cmd/api

## db/psql: connecting to database using docker
.PHONY: db/psql
db/psql:
	docker exec -it postgres14 psql ${GREENLIGHT_DB_DSN}

## db/migrations/new name=$1: create a new database migration
.PHONY: db/migration/new
db/migration/new:
	@echo 'create migrations file for ${name}'
	migrate create -seq -ext=sql -dir=./migrations ${name}

## db/migrations/up apply all up database migrations
.PHONY: db/migration/up
db/migration/up: confirm
	@echo 'running migrations...'
	migrate -path=./migrations -database=${GREENLIGHT_DB_DSN} up

# ==================================================================================== # 
# QUALITY CONTROL
# ==================================================================================== #

## tidy: tidy dependencies and format, vet and test all code
.PHONY: audit
audit: vendor
	@echo 'Formatting code...'
	go fmt ./...
	@echo 'Vetting code...'
	go vet ./...
	staticcheck ./...
	@echo 'Running tests...'
	go test -race -vet=off ./...

## vendor: tidy and vendor dependencies
.PHONY: vendor
vendor:
	@echo 'Tidying and verifying module dependencies...'
	go mod tidy
	go mod verify
	@echo 'Vendoring dependencies...'
	go mod vendor