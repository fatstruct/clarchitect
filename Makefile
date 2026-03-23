VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -X github.com/fatstruct/clarchitect/internal/version.Version=$(VERSION)
BINARY  := clarchitect

.PHONY: build test vet lint clean install

build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) .

test:
	go test ./...

vet:
	go vet ./...

lint: vet

clean:
	rm -f $(BINARY)

install:
	go install -ldflags "$(LDFLAGS)" .
