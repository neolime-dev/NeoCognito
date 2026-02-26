# NeoCognito Makefile
# Automates build and installation for Linux systems.

BINARY_NAME=neocognito
INSTALL_PATH=/usr/bin/$(BINARY_NAME)

.PHONY: all build install uninstall-old check-path clean test lint release tag help

all: build

VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "v1.0.0-dev")
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_DATE=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
LDFLAGS=-X 'github.com/neolime-dev/neocognito/internal/info.Version=$(VERSION)' \
        -X 'github.com/neolime-dev/neocognito/internal/info.Commit=$(COMMIT)' \
        -X 'github.com/neolime-dev/neocognito/internal/info.BuildDate=$(BUILD_DATE)'

build:
	@echo "Building NeoCognito $(VERSION)..."
	go build -ldflags "$(LDFLAGS)" -o $(BINARY_NAME) main.go
	@echo "Build complete: ./$(BINARY_NAME)"

install: build check-path
	@echo "Installing to $(INSTALL_PATH)..."
	sudo cp ./$(BINARY_NAME) $(INSTALL_PATH)
	@echo "Installation successful! You can now run '$(BINARY_NAME)' from anywhere."

uninstall-old:
	@echo "Removing rogue binaries from other locations..."
	rm -f $(HOME)/.local/bin/$(BINARY_NAME)
	rm -f $(HOME)/go/bin/$(BINARY_NAME)
	@echo "Cleanup complete."

check-path:
	@echo "Checking for shadowing binaries..."
	@if [ "$$(which $(BINARY_NAME) 2>/dev/null)" != "$(INSTALL_PATH)" ] && [ -n "$$(which $(BINARY_NAME) 2>/dev/null)" ]; then \
		echo "WARNING: Another version of $(BINARY_NAME) found at $$(which $(BINARY_NAME))"; \
		echo "This will likely shadow the version at $(INSTALL_PATH)."; \
		echo "Run 'make uninstall-old' to clean up rogue binaries."; \
	fi

clean:
	@echo "Cleaning up..."
	rm -f $(BINARY_NAME)
	@echo "Binary removed."

test:
	@echo "Running tests..."
	go test -race -timeout 120s ./...

lint:
	@echo "Running go vet..."
	go vet ./...
	@echo "Running staticcheck..."
	@if command -v staticcheck >/dev/null 2>&1; then \
		staticcheck ./...; \
	else \
		echo "staticcheck not installed; run: go install honnef.co/go/tools/cmd/staticcheck@latest"; \
	fi

release:
	@if ! command -v goreleaser >/dev/null 2>&1; then \
		echo "goreleaser not installed; run: go install github.com/goreleaser/goreleaser/v2@latest"; \
		exit 1; \
	fi
	goreleaser release --clean

tag:
	@if [ -z "$(VERSION)" ]; then \
		echo "Usage: make tag VERSION=v1.2.3"; \
		exit 1; \
	fi
	git tag -a $(VERSION) -m "Release $(VERSION)"
	git push origin $(VERSION)
	@echo "Tag $(VERSION) pushed. The release workflow will build and publish automatically."

help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  build    - Compile the project (creates local ./neocognito)"
	@echo "  install  - Compile and copy to $(INSTALL_PATH) (requires sudo)"
	@echo "  test     - Run all tests with race detector"
	@echo "  lint     - Run go vet and staticcheck"
	@echo "  clean    - Remove the local binary"
	@echo "  tag      - Create and push a release tag  (make tag VERSION=v1.2.3)"
	@echo "  release  - Build and publish a release locally with goreleaser"
	@echo "  help     - Show this help message"
