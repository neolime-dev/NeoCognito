# NeoCognito Makefile
# Automates build and installation for Linux systems.

BINARY_NAME=neocognito
INSTALL_PATH=/usr/bin/$(BINARY_NAME)

.PHONY: all build install clean help

all: build

build:
	@echo "🛠️  Building NeoCognito..."
	go build -o $(BINARY_NAME) main.go
	@echo "✅ Build complete: ./$(BINARY_NAME)"

install: build
	@echo "🚀 Installing to $(INSTALL_PATH)..."
	sudo cp ./$(BINARY_NAME) $(INSTALL_PATH)
	@echo "✨ Installation successful! You can now run '$(BINARY_NAME)' from anywhere."

clean:
	@echo "🧹 Cleaning up..."
	rm -f $(BINARY_NAME)
	@echo "🗑️  Binary removed."

help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  build   - Compile the project (creates local ./neocognito)"
	@echo "  install - Compile and copy to $(INSTALL_PATH) (requires sudo)"
	@echo "  clean   - Remove the local binary"
	@echo "  help    - Show this help message"
