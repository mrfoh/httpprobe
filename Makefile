BINARY_NAME=httpprobe

.PHONY: test lint build build-linux build-linux-arm build-macos build-macos-arm build-macos-universal prepare clean

.DEFAULT_GOAL := build

prepare:
	@echo "Preparing build environment..."
	@mkdir -p bin

test:
	@echo "Running tests..."
	@go test -v ./...

lint:
	@echo "Running linter..."
	@golangci-lint run

build: prepare
	@echo "Building for local environment..."
	CGO_ENABLED=0 go build -o bin/$(BINARY_NAME) cmd/main.go

build-linux: prepare
	@echo "Building for Linux (AMD64)..."
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/$(BINARY_NAME)_linux_amd64 cmd/main.go

build-linux-arm: prepare
	@echo "Building for Linux ARM (ARM64)..."
	@CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o bin/$(BINARY_NAME)_linux_arm64 cmd/main.go

build-macos: prepare
	@echo "Building for macOS (AMD64)..."
	@CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o bin/$(BINARY_NAME)_darwin cmd/main.go

build-macos-arm: prepare
	@echo "Building for macOS (ARM64)..."
	@CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o bin/$(BINARY_NAME)_apple_silicon cmd/main.go

build-macos-universal: build-macos build-macos-arm
	@echo "Creating universal binary for macOS..."
	@lipo -create -output bin/$(BINARY_NAME)_darwin_universal \
		bin/$(BINARY_NAME)_darwin bin/$(BINARY_NAME)_apple_silicon

clean:
	@echo "Cleaning build environment..."
	@rm -rf bin
