BINARY := zocli
PKG := ./cmd/zocli
BIN_DIR := bin
VERSION ?= dev

.PHONY: build
build:
	@mkdir -p $(BIN_DIR)
	go build -ldflags "-X main.version=$(VERSION)" -o $(BIN_DIR)/$(BINARY) $(PKG)

.PHONY: run
run:
	go run $(PKG)

.PHONY: fmt
fmt:
	gofmt -w cmd/zocli internal

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: clean
clean:
	@rm -rf $(BIN_DIR)

.PHONY: test
test:
	go test -v ./...
