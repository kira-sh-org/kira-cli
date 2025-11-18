.PHONY: build test test-coverage clean clean-all install uninstall lint fmt check build-all release-snapshot install-tools clean-tools dev-setup help demo

PREFIX ?= /usr/local
DESTDIR ?=

# Build the kira binary
build:
	$(eval GIT_TAG := $(shell git describe --tags --always 2>/dev/null || echo dev))
	$(eval GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown))
	$(eval BUILD_DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ))
	$(eval GIT_DIRTY := $(shell test -n "$(shell git status --porcelain 2>/dev/null)" && echo dirty || echo clean))
	go build -ldflags "-X 'kira/internal/commands.Version=$(GIT_TAG)' -X 'kira/internal/commands.Commit=$(GIT_COMMIT)' -X 'kira/internal/commands.BuildDate=$(BUILD_DATE)' -X 'kira/internal/commands.Dirty=$(GIT_DIRTY)'" -o kira cmd/kira/main.go

# Run tests
test:
	go test ./...

# Run tests with coverage
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	rm -f kira coverage.out coverage.html kira-*-amd64 kira-*-amd64.exe
	rm -rf dist/

# Clean everything: build artifacts, dev tools, and installed binary
clean-all: clean clean-tools uninstall
	@echo "All cleanup complete"

# Install kira to $(PREFIX)/bin
install: build
	install -d "$(DESTDIR)$(PREFIX)/bin"
	install -m 0755 kira "$(DESTDIR)$(PREFIX)/bin/kira"

# Uninstall kira from $(PREFIX)/bin
uninstall:
	@echo "Uninstalling kira from $(DESTDIR)$(PREFIX)/bin..."
	@rm -f "$(DESTDIR)$(PREFIX)/bin/kira"
	@echo "kira uninstalled"

# Run linter
lint:
	golangci-lint run

# Format code (writes changes) via golangci-lint config
fmt:
	golangci-lint run --fix

# Run all checks (lint includes formatting and vet checks via golangci-lint)
check: lint test

# Build for multiple platforms
build-all:
	GOOS=linux GOARCH=amd64 go build -o kira-linux-amd64 cmd/kira/main.go
	GOOS=darwin GOARCH=amd64 go build -o kira-darwin-amd64 cmd/kira/main.go
	GOOS=windows GOARCH=amd64 go build -o kira-windows-amd64.exe cmd/kira/main.go

# Test release locally using GoReleaser (does not create GitHub release)
release-snapshot:
	@if ! command -v goreleaser >/dev/null 2>&1; then \
		echo "GoReleaser not found. Install it with: go install github.com/goreleaser/goreleaser@latest"; \
		exit 1; \
	fi
	goreleaser release --snapshot --clean

# Install required developer tools
install-tools:
	@echo "Installing developer tools..."
	@BIN_DIR=$$(go env GOPATH)/bin; \
	  echo "Checking if golangci-lint is already installed..."; \
	  if command -v golangci-lint >/dev/null 2>&1; then \
	    echo "golangci-lint found, checking version..."; \
	    INSTALLED_VERSION=$$(golangci-lint version --format short 2>/dev/null | grep -oE 'v[0-9]+\.[0-9]+\.[0-9]+' | head -1 || echo ""); \
	    if [ "$$INSTALLED_VERSION" = "$(GOLANGCI_LINT_VERSION)" ]; then \
	      echo "golangci-lint $(GOLANGCI_LINT_VERSION) already installed, skipping download"; \
	      exit 0; \
	    else \
	      echo "golangci-lint $$INSTALLED_VERSION installed, upgrading to $(GOLANGCI_LINT_VERSION)"; \
	    fi; \
	  fi; \
	  echo "Installing golangci-lint $(GOLANGCI_LINT_VERSION) to $$BIN_DIR"; \
	  echo "Downloading install script to temporary file (safer than piping curl to sh)..."; \
	  INSTALL_SCRIPT=$$(mktemp) || exit 1; \
	  curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh -o $$INSTALL_SCRIPT || (rm -f $$INSTALL_SCRIPT; exit 1); \
	  echo "Running install script (it handles SHA256 checksum verification)..."; \
	  sh $$INSTALL_SCRIPT -b $$BIN_DIR $(GOLANGCI_LINT_VERSION); \
	  echo "Cleaning up temporary install script..."; \
	  rm -f $$INSTALL_SCRIPT; \
	  echo ""; \
	  echo "Installing GoReleaser..."; \
	  if command -v goreleaser >/dev/null 2>&1; then \
	    echo "GoReleaser already installed, skipping"; \
	  else \
	    go install github.com/goreleaser/goreleaser@latest || echo "Failed to install GoReleaser"; \
	  fi

# Clean up developer tools installed by install-tools/dev-setup
clean-tools:
	@echo "Cleaning up developer tools..."
	@BIN_DIR=$$(go env GOPATH)/bin; \
	  if [ -f "$$BIN_DIR/golangci-lint" ]; then \
	    echo "Removing golangci-lint from $$BIN_DIR..."; \
	    rm -f "$$BIN_DIR/golangci-lint"; \
	    echo "golangci-lint removed"; \
	  else \
	    echo "golangci-lint not found in $$BIN_DIR"; \
	  fi; \
	  if [ -f "$$BIN_DIR/goreleaser" ]; then \
	    echo "Removing goreleaser from $$BIN_DIR..."; \
	    rm -f "$$BIN_DIR/goreleaser"; \
	    echo "goreleaser removed"; \
	  else \
	    echo "goreleaser not found in $$BIN_DIR"; \
	  fi
	@echo "Developer tools cleaned up"

# Development setup
GOLANGCI_LINT_VERSION ?= v1.55.2

dev-setup: install-tools
	go mod download
	go mod tidy

# Run kira with help
help:
	./kira --help

# Demo initialization
demo:
	./kira init demo-workspace
	cd demo-workspace && ../kira new prd "Demo Feature" todo "This is a demo feature"
	cd demo-workspace && ../kira move 001 doing
	cd demo-workspace && ../kira save "Initial demo setup"

