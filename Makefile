APP_NAME := zpcli
DIST_DIR := dist
VERSION := $(shell cat VERSION)
COMMIT := $(shell git rev-parse --short HEAD)
BUILD_DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -X 'zpcli/internal/buildinfo.Version=$(VERSION)' -X 'zpcli/internal/buildinfo.Commit=$(COMMIT)' -X 'zpcli/internal/buildinfo.BuildDate=$(BUILD_DATE)'

.PHONY: build build-linux-amd64 build-linux-arm64 build-linux release-artifacts clean test

build:
	mkdir -p bin
	go build -ldflags "$(LDFLAGS)" -o bin/$(APP_NAME) .

build-linux-amd64:
	mkdir -p $(DIST_DIR)
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(APP_NAME)_linux_amd64 .

build-linux-arm64:
	mkdir -p $(DIST_DIR)
	GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(APP_NAME)_linux_arm64 .

build-linux: build-linux-amd64 build-linux-arm64

release-artifacts: build-linux
	cd $(DIST_DIR) && shasum -a 256 $(APP_NAME)_linux_amd64 $(APP_NAME)_linux_arm64 > checksums.txt

clean:
	rm -rf bin $(DIST_DIR)

test:
	go test ./...
