SHELL := /bin/bash

# The name of the executable (default is current directory name)
TARGET := $(echo $${PWD\#\#*/})
.DEFAULT_GOAL: $(TARGET)

# These will be provided to the target
VERSION := `git describe --tags`
BUILD := `git rev-parse --short HEAD`

# Use linker flags to provide version/build settings to the target
LDFLAGS=-ldflags "-X=main.Version=$(VERSION) -X=main.Build=$(BUILD) -s -w"

# go source files, ignore vendor directory
SRC = $(find . -type f -name '*.go' -not -path "./vendor/*")

.PHONY: all build clean install uninstall fmt check run doc

all: check install

$(TARGET): $(SRC)
	@go build $(LDFLAGS) -o $(TARGET)

build: $(TARGET)
	@true

clean:
	@rm -vf $(TARGET)

install:
	@go install $(LDFLAGS)

uninstall: clean
	@echo rm -vf $$(which $(TARGET))

fmt:
	@go fmt $(SRC)

check:
	@mkdir -p tests
	@go test $(LDFLAGS)
	@go test -race ./...
	@go test -coverprofile=tests/coverage.out ./...
	@go tool cover -func=tests/coverage.out
	@go tool cover -html=tests/coverage.out -o tests/coverage.html
	@rm -f tests/coverage.out
	@go vet ${SRC}

	@echo "Coverage file:"
	@echo " - $(PWD)/tests/coverage.html"

run: install
	@$(TARGET)

doc:
	@echo "--> Wait a few seconds and visit http://localhost:6060/pkg/github.com/ritchies/ctftool to see the documentation"
	@godoc -http=:6060 -index