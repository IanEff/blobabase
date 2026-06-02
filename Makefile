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
	go build -o blobabase .

run:
	go run .

tidy:
	go mod tidy

clean:
	rm -f blobabase
