SHELL := /bin/bash

# The name of the executable
TARGET 		?= $(shell basename `go list`)

# These will be provided to the target
VERSION 	?= $(shell git describe --tags --always)
BUILD 		?= $(shell git rev-parse --short HEAD)
CURDATE 	?= $(shell date +%Y/%m/%d_%H:%M:%S)

# Source files, ignore vendor directory
SRC 		?= $(shell find . -type f -name '*.go' -not -path "./vendor/*")

# Use linker flags to provide version/build settings to the target
LDFLAGS=-ldflags "-X=github.com/ritchies/ctftool/cmd.Version=$(VERSION) -X=github.com/ritchies/ctftool/cmd.Commit=$(BUILD) -X=github.com/ritchies/ctftool/cmd.BuildTime=$(CURDATE) -s -w"

.DEFAULT_GOAL: $(TARGET)
.PHONY: help build install uninstall clean fmt vet check test test-it test-bench test-race test-cover test-all run doc all

default: help

help: ## show this help
	@echo 'usage: make [target] ...'
	@echo ''
	@echo 'targets:'
	@egrep '^(.+)\:\ .*##\ (.+)' ${MAKEFILE_LIST} | sed 's/:.*##/#/' | column -t -c 2 -s '#'

$(TARGET): $(SRC)
	@go build $(LDFLAGS) -o bin/$(TARGET)
	@echo "Built $(TARGET) version $(VERSION) commit $(BUILD)"
	@echo "Built $(TARGET) version $(VERSION) commit $(BUILD)" > bin/$(TARGET).version
	@echo "You can run the program by typing './bin/$(TARGET)'"

all: clean fmt test build ## clean, format, unit test and build

build: $(TARGET) ## build the executable
	@true

install: ## install the executable to $GOPATH/bin	
	@go install -v $(LDFLAGS) ./...
	@echo "Installed $(TARGET) version $(VERSION) commit $(BUILD)"
	@echo "You can run the program by typing '$(TARGET)'"

uninstall: clean ## uninstall the executable from $GOPATH/bin
	go clean -i ./...
	@rm -vf `which $(TARGET)`

clean: ## remove all generated files
	go clean
	@rm -vf `which $(TARGET)`
	@rm -vf bin/$(TARGET) bin/$(TARGET).version
	@rm -vf tests/coverage.out tests/coverage.html
	@rm -vrf tests bin output

fmt: ## format the source files
	go fmt ./...

vet: ## run go vet on the source files
	go vet ./...

tidy: ## go mod tidy on the source files
	go mod tidy

check: fmt vet tidy
	go test ./... -short

test: vet ## run short unit tests
	go test -v ./... -short

test-it: ## run the integration tests
	go test -v ./...

test-bench: ## run the benchmark tests
	go test -bench ./...

test-race: ## run the race condition tests
	go test -race ./...

test-cover: ## generate test coverage report
	@rm -vrf tests
	@mkdir -p tests
	@go test -coverprofile=tests/coverage.out ./...
	@go tool cover -func=tests/coverage.out
	@go tool cover -html=tests/coverage.out -o tests/coverage.html

	@echo "Coverage file:"
	@echo " - $(PWD)/tests/coverage.html"
	@rm -f tests/coverage.out

test-all: test test-it test-bench test-race test-cover ## run all tests

run: install ## install and run the binary
	@$(TARGET)

doc: ## generate docs
	@echo "NOTE: Visit http://localhost:6060/pkg/github.com/ritchies/ctftool to see the documentation"
	godoc -http=:6060 -index