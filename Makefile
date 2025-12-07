.PHONY: all build run clean test lint fmt tidy

BINARY := net-tui
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-s -w -X main.version=$(VERSION)"

all: build

build:
	go build $(LDFLAGS) -o $(BINARY) .

run: build
	./$(BINARY)

# sudo for full process visibility
run-root: build
	sudo ./$(BINARY)

clean:
	rm -f $(BINARY)
	go clean

test:
	go test -v -race ./...

lint:
	golangci-lint run

fmt:
	go fmt ./...
	goimports -w .

tidy:
	go mod tidy

# dev: rebuild on file changes (requires entr)
dev:
	find . -name '*.go' | entr -r make run

# show binary size
size: build
	ls -lh $(BINARY)
	file $(BINARY)
