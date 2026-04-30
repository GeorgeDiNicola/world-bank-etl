BINARY_NAME=world-bank-etl
MAIN_PACKAGE_PATH=./cmd/wb/main.go
GIT_COMMIT ?= $(shell git rev-parse --short HEAD)
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)

LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.Commit=${GIT_COMMIT}"

.PHONY: test
test: ## Run unit tests
	go test -v -race -cover ./...

.PHONY: tidy
tidy: ## Download dependencies and remove unused ones
	go mod tidy
	go mod verify

.PHONY: build
build: tidy ## Build the binary for the current architecture
	go build ${LDFLAGS} -o bin/${BINARY_NAME} ${MAIN_PACKAGE_PATH}

.PHONY: build-all
build-all: ## Build for Linux, Windows, and macOS
	GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o bin/${BINARY_NAME}-linux-amd64 ${MAIN_PACKAGE_PATH}
	GOOS=windows GOARCH=amd64 go build ${LDFLAGS} -o bin/${BINARY_NAME}-windows-amd64.exe ${MAIN_PACKAGE_PATH}
	GOOS=darwin GOARCH=arm64 go build ${LDFLAGS} -o bin/${BINARY_NAME}-darwin-arm64 ${MAIN_PACKAGE_PATH}

.PHONY: clean
clean: ## Remove build artifacts
	rm -rf bin/

PHONY: help
help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'