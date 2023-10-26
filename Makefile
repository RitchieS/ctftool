# -----------------------------------------------------------------------------
# Variables
# -----------------------------------------------------------------------------
SHELL := /bin/bash

# Application Info
TARGET  ?= $(shell basename `go list`)
VERSION ?= $(shell git describe --tags --always)
BUILD   ?= $(shell git rev-parse --short HEAD)
CURDATE ?= $(shell date +%Y/%m/%d_%H:%M:%S)

# Source files
SRC ?= $(shell find . -type f -name '*.go' -not -path "./vendor/*")

# Linker Flags
LDFLAGS = -ldflags "-X=main.version=$(VERSION) -X=main.commit=$(BUILD) -X=main.date=$(CURDATE) -s -w"

# -----------------------------------------------------------------------------
# Targets
# -----------------------------------------------------------------------------
.DEFAULT_GOAL := help

.PHONY: help
help:  ## Display help
	@echo 'Usage: make [target] ...'
	@echo ''
	@echo 'Targets:'
	@egrep '^(.+)\:\ .*##\ (.+)' ${MAKEFILE_LIST} | sed 's/:.*##/#/' | column -t -c 2 -s '#'

$(TARGET): $(SRC)
	@go build $(LDFLAGS) -o bin/$(TARGET)
	@echo "Built $(TARGET) version $(VERSION) commit $(BUILD)" > bin/$(TARGET).version
	@echo "You can run the program by typing './bin/$(TARGET)'"

.PHONY: all
all: clean fmt test build ## clean, format, unit test and build

build: $(TARGET) ## Go: build executable
	@true

.PHONY: build-single
build-single:  ## Build a single target using GoReleaser
	@goreleaser build --rm-dist --single-target

.PHONY: build-all
build-all:  ## Build for all platforms using GoReleaser
	@goreleaser build --rm-dist 

.PHONY: build-tools
build-tools:  ## Install required tools
	go install -v github.com/goreleaser/goreleaser@latest
	go install -v golang.org/x/tools/cmd/godoc@latest
	go install -v github.com/golangci/golangci-lint/cmd/golangci-lint@v1.46.2
	go install -v github.com/caarlos0/svu@latest

.PHONY: release
release:  ## Create a release using GoReleaser
	@goreleaser release --rm-dist

.PHONY: major minor patch
major minor patch:
	git tag $$(svu $@)

.PHONY: install
install:  ## Install the executable
	@go install -v $(LDFLAGS) ./...
	@echo "Installed $(TARGET) version $(VERSION) commit $(BUILD)"
	@echo "You can run the program by typing '$(TARGET)'"

.PHONY: uninstall
uninstall: clean  ## Uninstall the executable
	go clean -i ./...
	@rm -vf `which $(TARGET)`

.PHONY: clean
clean:  ## Clean up generated files
	go clean
	@rm -rf bin/ dist/ tests/ output `which $(TARGET)`

.PHONY: fmt
fmt:  ## Format the source code
	gofmt -s -w $(SRC)

.PHONY: vet
vet:  ## Run go vet on the source code
	go vet ./...

.PHONY: lint
lint:  ## Run golangci-lint on the source code
	golangci-lint run --exclude-use-default ./...

.PHONY: tidy
tidy:  ## Clean up Go modules
	go mod tidy

.PHONY: check
check: fmt vet tidy  ## Run code checks and short tests
	go test ./... -short
	goreleaser check

.PHONY: test test-it test-bench test-race test-cover test-all
test: vet  ## Run short unit tests
	go test -v ./... -short

test-it:  ## Run integration tests
	go test -v ./...

test-bench:  ## Run benchmarks
	go test -bench ./...

test-race:  ## Run race condition tests
	go test -race ./...

test-cover:  ## Generate test coverage report
	@mkdir -p tests
	@go test -coverprofile=tests/coverage.out ./...
	@go tool cover -func=tests/coverage.out
	@go tool cover -html=tests/coverage.out -o tests/coverage.html
	@echo "Coverage file:"
	@echo " - $(PWD)/tests/coverage.html"
	@rm -f tests/coverage.out

test-all: test test-it test-bench test-race test-cover  ## Run all tests

.PHONY: run
run: install  ## Install and run the executable
	@$(TARGET)

.PHONY: doc
doc:  ## Generate documentation
	@echo "NOTE: Visit http://localhost:6060/pkg/github.com/ritchies/ctftool to see the documentation"
	godoc -http=:6060 -index