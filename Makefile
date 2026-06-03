PROJECT := blobabase

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT  := $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE    := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

.PHONY: all ci fmt vet test build run tidy clean

all: build

ci: fmt vet test build

fmt:
	go fmt ./...

vet:
	go vet ./...

test:
	go test -race ./...

build:
	go build $(LDFLAGS) -o ./bin/$(PROJECT) .

run:
	go run .

tidy:
	go mod tidy

clean:
	rm -f bin/blobabase
