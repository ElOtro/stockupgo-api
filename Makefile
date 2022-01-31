## help: print this help message
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

## run: run the cmd/api application
run:
	go run ./cmd/api

## build: build the cmd/api application
build:
	@echo 'Building cmd/api...'
	go build -ldflags='-s' -o=./bin/api ./cmd/api

## migrations/create name=$1: create a new database migration
migration/create:
	@echo 'Creating migration files for ${name}...'
	migrate create -seq -ext=.sql -dir=./migrations ${name}

## migrations/up: apply all up database migrations
migrations/up:
	@echo 'Running up migrations...'
	migrate -path ./migrations -database "postgres://elotro@localhost/stockup_dev?sslmode=disable" up

## migrations/down: apply all down database migrations
migrations/down:
	@echo 'Running up migrations...'
	migrate -path ./migrations -database "postgres://elotro@localhost/stockup_dev?sslmode=disable" down