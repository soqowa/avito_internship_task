APP_NAME=reviewer-svc

GOFILES := $(shell find . -name '*.go' -not -path './vendor/*')

.PHONY: all build run test lint gen compose-up compose-down migrate-up migrate-down

all: build

build:
	go build -o bin/$(APP_NAME) ./cmd/reviewer-svc

run: build
	PORT=8080 ./bin/$(APP_NAME)

test:
	go test ./...

lint:
	golangci-lint run ./...

gen:
		 "OpenAPI codegen no longer used"

compose-up:
	 docker-compose up --build

compose-down:
	 docker-compose down

migrate-up:
	goose -dir ./migrations postgres "$$DB_DSN" up

migrate-down:
	goose -dir ./migrations postgres "$$DB_DSN" down

