APP      := skwad
MODULE   := github.com/Jared-Boschmann/skwad-linux
BIN      := ./bin/$(APP)
CMD      := ./cmd/$(APP)
VERSION  := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS  := -ldflags "-s -w -X main.version=$(VERSION)"

PKG_DEPS := vte-2.91 gtk+-3.0

.PHONY: all build run test clean deps install

all: build

## build: compile the application
build:
	go build $(LDFLAGS) -o $(BIN) $(CMD)

## run: run without building a binary
run:
	go run $(LDFLAGS) $(CMD)

## install: build and install to /usr/local/bin
install: build
	install -m 755 $(BIN) /usr/local/bin/$(APP)

## test: run all tests
test:
	go test ./...

## clean: remove build artifacts
clean:
	rm -rf ./bin

## deps: verify CGo pkg-config dependencies are available
deps:
	@pkg-config --exists $(PKG_DEPS) || \
		(echo "ERROR: missing pkg-config deps: $(PKG_DEPS). Install libvte-2.91-dev and libgtk-3-dev" && exit 1)

## tidy: tidy Go modules
tidy:
	go mod tidy

## fmt: format Go source
fmt:
	gofmt -w -s .

## vet: run go vet
vet:
	go vet ./...

## lint: run golangci-lint if available
lint:
	@which golangci-lint > /dev/null && golangci-lint run || echo "golangci-lint not installed"
