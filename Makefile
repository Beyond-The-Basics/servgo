BINARY   := servgo
CMD_DIR  := ./cmd
BIN_DIR  := bin

GO       := go
GOFLAGS  :=

.PHONY: all build run test clean vet fmt lint help

all: build

## build: compile the binary into bin/
build:
	@mkdir -p $(BIN_DIR)
	$(GO) build $(GOFLAGS) -o $(BIN_DIR)/$(BINARY) $(CMD_DIR)

## run: build and execute the server
run: build
	./$(BIN_DIR)/$(BINARY)

## test: run all tests
test:
	$(GO) test ./tests/... -v -timeout 60s

## vet: run go vet on all packages
vet:
	$(GO) vet ./...

## fmt: format all Go source files
fmt:
	$(GO) fmt ./...

## lint: run staticcheck (install with: go install honnef.co/go/tools/cmd/staticcheck@latest)
lint:
	staticcheck ./...

## clean: remove build artifacts
clean:
	@rm -rf $(BIN_DIR)

## help: list available targets
help:
	@grep -E '^##' $(MAKEFILE_LIST) | sed 's/## /  /'
