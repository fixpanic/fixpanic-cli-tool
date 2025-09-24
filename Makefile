# Fixpanic CLI Makefile

# Variables
BINARY_NAME=fixpanic
VERSION?=dev
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

# Go commands
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt
GOVET=$(GOCMD) vet

# Directories
BUILD_DIR=build
RELEASE_DIR=release
SCRIPTS_DIR=scripts

# Platforms
PLATFORMS=linux/amd64 linux/arm64 linux/386 linux/arm darwin/amd64 darwin/arm64 windows/amd64

# Default target
.PHONY: all
all: clean build

# Build for current platform
.PHONY: build
build:
	@echo "Building $(BINARY_NAME) for current platform..."
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) main.go
	@echo "Build complete: $(BINARY_NAME)"

# Build for all platforms
.PHONY: build-all
build-all: clean
	@echo "Building for all platforms..."
	@mkdir -p $(BUILD_DIR)
	@for platform in $(PLATFORMS); do \
		GOOS=$${platform%/*} GOARCH=$${platform#*/} \
		$(GOBUILD) $(LDFLAGS) \
		-o $(BUILD_DIR)/$(BINARY_NAME)-$${platform%/*}-$${platform#*/}$(if $(findstring windows,$${platform}),.exe,) \
		main.go; \
		echo "Built: $(BUILD_DIR)/$(BINARY_NAME)-$${platform%/*}-$${platform#*/}$(if $(findstring windows,$${platform}),.exe,)"; \
	done
	@echo "All builds complete!"

# Create release packages
.PHONY: release
release: build-all
	@echo "Creating release packages..."
	@mkdir -p $(RELEASE_DIR)
	@for platform in $(PLATFORMS); do \
		OS=$${platform%/*}; \
		ARCH=$${platform#*/}; \
		BINARY=$(BINARY_NAME)-$${OS}-$${ARCH}$$(if [ "$$OS" = "windows" ]; then echo ".exe"; fi); \
		if [ "$$OS" != "windows" ]; then \
			tar -czf $(RELEASE_DIR)/$$BINARY.tar.gz -C $(BUILD_DIR) $$BINARY; \
			echo "Created: $(RELEASE_DIR)/$$BINARY.tar.gz"; \
		else \
			cp $(BUILD_DIR)/$$BINARY $(RELEASE_DIR)/$$BINARY; \
			echo "Created: $(RELEASE_DIR)/$$BINARY"; \
		fi; \
	done
	@echo "Release packages created in $(RELEASE_DIR)/"

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Format code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	$(GOFMT) ./...

# Vet code
.PHONY: vet
vet:
	@echo "Vetting code..."
	$(GOVET) ./...

# Lint code
.PHONY: lint
lint:
	@echo "Linting code..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found, installing..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
		golangci-lint run; \
	fi

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR) $(RELEASE_DIR) $(BINARY_NAME)
	@echo "Clean complete"

# Install dependencies
.PHONY: deps
deps:
	@echo "Installing dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy
	@echo "Dependencies installed"

# Update dependencies
.PHONY: deps-update
deps-update:
	@echo "Updating dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy
	$(GOGET) -u ./...
	@echo "Dependencies updated"

# Run the CLI
.PHONY: run
run: build
	@echo "Running $(BINARY_NAME)..."
	./$(BINARY_NAME)

# Install locally
.PHONY: install
install: build
	@echo "Installing $(BINARY_NAME)..."
	sudo cp $(BINARY_NAME) /usr/local/bin/
	sudo chmod +x /usr/local/bin/$(BINARY_NAME)
	@echo "Installed to /usr/local/bin/$(BINARY_NAME)"

# Uninstall locally
.PHONY: uninstall
uninstall:
	@echo "Uninstalling $(BINARY_NAME)..."
	sudo rm -f /usr/local/bin/$(BINARY_NAME)
	@echo "Uninstalled"

# Docker build
.PHONY: docker-build
docker-build:
	@echo "Building Docker image..."
	docker build -t fixpanic/$(BINARY_NAME):$(VERSION) \
		--build-arg VERSION=$(VERSION) \
		--build-arg COMMIT=$(COMMIT) \
		--build-arg DATE=$(DATE) \
		.
	@echo "Docker image built: fixpanic/$(BINARY_NAME):$(VERSION)"

# Docker run
.PHONY: docker-run
docker-run: docker-build
	@echo "Running Docker container..."
	docker run --rm -it fixpanic/$(BINARY_NAME):$(VERSION)

# Help
.PHONY: help
help:
	@echo "Fixpanic CLI Makefile"
	@echo ""
	@echo "Available targets:"
	@echo "  make build         - Build for current platform"
	@echo "  make build-all     - Build for all platforms"
	@echo "  make release       - Create release packages"
	@echo "  make test          - Run tests"
	@echo "  make test-coverage - Run tests with coverage"
	@echo "  make fmt           - Format code"
	@echo "  make vet           - Vet code"
	@echo "  make lint          - Lint code"
	@echo "  make clean         - Clean build artifacts"
	@echo "  make deps          - Install dependencies"
	@echo "  make deps-update   - Update dependencies"
	@echo "  make run           - Build and run the CLI"
	@echo "  make install       - Install locally"
	@echo "  make uninstall     - Uninstall locally"
	@echo "  make docker-build  - Build Docker image"
	@echo "  make docker-run    - Run Docker container"
	@echo "  make help          - Show this help"

# Default help
.PHONY: default
default: help