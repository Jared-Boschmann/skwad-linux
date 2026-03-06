APP      := skwad
MODULE   := github.com/kochava-studios/skwad-linux
BIN      := ./bin/$(APP)
CMD      := ./cmd/$(APP)

PKG_DEPS := vte-2.91 gtk+-3.0

.PHONY: all build run test clean deps

all: build

## build: compile the application
build: deps
	go build -o $(BIN) $(CMD)

## run: build and run the application
run: build
	$(BIN)

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
