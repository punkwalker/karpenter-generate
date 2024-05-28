COMMIT = $(shell git rev-parse --short HEAD)
DATE = $(shell date -u "+%FT%TZ")
VERSION = $(shell git describe --tags --always --dirty=dev)
VERSION_PKG := github.com/punkwalker/karpenter-generate/cmd

LDFLAGS += -s -w
LDFLAGS += -X $(VERSION_PKG).date=$(DATE)
LDFLAGS += -X $(VERSION_PKG).commit=$(COMMIT)
LDFLAGS += -X $(VERSION_PKG).version=$(VERSION)

build:
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)"

install:
	CGO_ENABLED=0 go install -ldflags "$(LDFLAGS)"

lint:
	golangci-lint run

tidy:
	go mod tidy

test:
	go test ./... -coverprofile=cover.out

coverage: 
	go tool cover -html=cover.out

.PHONY: build install lint tidy test coverage


