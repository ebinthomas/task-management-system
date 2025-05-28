.PHONY: all build test clean run docker-build docker-run

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME=task-management-system
BINARY_UNIX=$(BINARY_NAME)_unix

all: test build

build:
	$(GOBUILD) -o bin/$(BINARY_NAME) -v cmd/api/main.go

test:
	$(GOTEST) -v ./...

clean:
	$(GOCLEAN)
	rm -f bin/$(BINARY_NAME)
	rm -f bin/$(BINARY_UNIX)

run:
	$(GOBUILD) -o bin/$(BINARY_NAME) -v cmd/api/main.go
	./bin/$(BINARY_NAME)

docker-build:
	docker-compose build

docker-run:
	docker-compose up

docker-down:
	docker-compose down

generate-token:
	go run cmd/tools/token_gen.go

lint:
	golangci-lint run

test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

deps:
	$(GOGET) -v ./...
	go mod tidy

migrate:
	go run cmd/tools/migrate.go

# Cross compilation
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o bin/$(BINARY_UNIX) -v cmd/api/main.go 