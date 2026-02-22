# NeoCognito Makefile
# Automates build and installation for Linux systems.

BINARY_NAME=neocognito
INSTALL_PATH=/usr/bin/$(BINARY_NAME)

.PHONY: all build install uninstall-old check-path clean help

all: build

VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "v1.0.0-dev")
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_DATE=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
LDFLAGS=-X 'github.com/lemondesk/neocognito/internal/info.Version=$(VERSION)' \
        -X 'github.com/lemondesk/neocognito/internal/info.Commit=$(COMMIT)' \
        -X 'github.com/lemondesk/neocognito/internal/info.BuildDate=$(BUILD_DATE)'

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

help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  build   - Compile the project (creates local ./neocognito)"
	@echo "  install - Compile and copy to $(INSTALL_PATH) (requires sudo)"
	@echo "  clean   - Remove the local binary"
	@echo "  help    - Show this help message"
