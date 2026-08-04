.PHONY: help build run dev test test-verbose clean install stop start restart lint fmt fmt-check vet assets assets-check check

# Default target
help:
	@echo "Blockbusterr - Makefile Commands"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Build Commands:"
	@echo "  build           Build the application binary"
	@echo "  install         Install dependencies"
	@echo "  clean           Remove build artifacts"
	@echo "  assets          Compile and vendor frontend assets"
	@echo "  assets-check    Verify committed frontend assets are current"
	@echo ""
	@echo "Run Commands:"
	@echo "  run             Build and run the application"
	@echo "  dev             Run the application (assumes already built)"
	@echo "  start           Start the application in background"
	@echo "  stop            Stop the application"
	@echo "  restart         Restart the application"
	@echo ""
	@echo "Test Commands:"
	@echo "  test            Run all tests"
	@echo "  test-verbose    Run all tests with verbose output"
	@echo "  test-coverage   Run tests with coverage report"
	@echo ""
	@echo "Code Quality:"
	@echo "  lint            Run linter"
	@echo "  fmt             Format code"
	@echo "  fmt-check       Verify Go formatting without changing files"
	@echo "  vet             Run go vet"
	@echo "  check           Run fmt, vet, and lint"
	@echo ""
	@echo "Docker Commands:"
	@echo "  docker-build    Build Docker image"
	@echo "  docker-run      Run Docker container"
	@echo "  beta-fast       Build and push beta (amd64 only - fast, use for testing)"
	@echo "  beta            Build and push beta (multi-platform - slow, use for releases)"
	@echo "  beta-test       Pull and run latest beta version locally"
	@echo ""

# Build the application
build:
	@echo "Building blockbusterr..."
	@go build -o bin/blockbusterr cmd/app/main.go
	@echo "✓ Build complete: bin/blockbusterr"

# Install dependencies
install:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy
	@echo "✓ Dependencies installed"

# Node is build-only; the compiled assets remain available to ordinary Go builds.
assets:
	@npm ci
	@npm run build:assets

assets-check:
	@npm ci
	@npm run check:assets

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@go clean
	@echo "✓ Clean complete"

# Build and run the application
run: build
	@echo "Starting blockbusterr..."
	@./bin/blockbusterr

# Run the application (assumes already built)
dev:
	go run cmd/app/main.go

# Start the application in background
start: build stop
	@echo "Starting blockbusterr in background..."
	@nohup ./bin/blockbusterr > blockbusterr.log 2>&1 &
	@sleep 2
	@if lsof -ti:9090 > /dev/null 2>&1; then \
		echo "✓ Blockbusterr started on http://127.0.0.1:9090"; \
		echo "  Logs: tail -f blockbusterr.log"; \
	else \
		echo "✗ Failed to start blockbusterr"; \
		tail -20 blockbusterr.log; \
		exit 1; \
	fi

# Stop the application
stop:
	@echo "Stopping blockbusterr..."
	@if lsof -ti:9090 > /dev/null 2>&1; then \
		lsof -ti:9090 | xargs kill -9; \
		echo "✓ Blockbusterr stopped"; \
	else \
		echo "✓ Blockbusterr not running"; \
	fi

# Restart the application
restart: stop start

# Run all tests
test:
	@echo "Running tests..."
	@go test ./... -v

# Run tests with verbose output
test-verbose:
	@echo "Running tests (verbose)..."
	@go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✓ Coverage report generated: coverage.html"

# Run tests for a specific package
test-pkg:
	@if [ -z "$(PKG)" ]; then \
		echo "Usage: make test-pkg PKG=<package>"; \
		echo "Example: make test-pkg PKG=./internal/scoring"; \
		exit 1; \
	fi
	@echo "Running tests for $(PKG)..."
	@go test -v $(PKG)

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@echo "✓ Code formatted"

fmt-check:
	@files="$$(find . -type f -name '*.go' -not -path './vendor/*' -print0 | xargs -0 gofmt -l)"; if [ -n "$$files" ]; then echo "Go files need formatting:"; echo "$$files"; exit 1; fi

# Run go vet
vet:
	@echo "Running go vet..."
	@go vet ./...
	@echo "✓ Vet complete"

# Run linter (requires golangci-lint)
lint:
	@echo "Running linter..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run ./... && \
		echo "✓ Lint complete"; \
	else \
		echo "golangci-lint is required"; \
		echo "Install: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		exit 1; \
	fi

# Run all checks
check: fmt-check vet lint
	@echo "✓ All checks passed"

# Build for production
build-prod:
	@echo "Building for production..."
	@CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o bin/blockbusterr cmd/app/main.go
	@echo "✓ Production build complete"

# Show application status
status:
	@if lsof -ti:9090 > /dev/null 2>&1; then \
		echo "✓ Blockbusterr is running"; \
		echo "  PID: $$(lsof -ti:9090)"; \
		echo "  URL: http://127.0.0.1:9090"; \
	else \
		echo "✗ Blockbusterr is not running"; \
	fi

# Show logs
logs:
	@if [ -f blockbusterr.log ]; then \
		tail -f blockbusterr.log; \
	else \
		echo "No log file found. Application may not be running in background."; \
	fi

# Docker commands (if using Docker)
docker-build:
	@echo "Building Docker image..."
	@docker build -t blockbusterr:latest .
	@echo "✓ Docker image built"

docker-run:
	@echo "Running Docker container..."
	@docker run -p 9090:9090 -v $(PWD)/data:/app/data -v $(PWD)/config:/app/config blockbusterr:latest

# Beta test - pull and run latest beta version locally
# Usage: make beta-test VERSION=v1.2.0
beta-test:
	@if [ -z "$(VERSION)" ]; then \
		echo "Usage: make beta-test VERSION=v1.2.0"; \
		echo "Example: make beta-test VERSION=v1.2.0"; \
		exit 1; \
	fi
	@echo "🔍 Finding latest beta version for $(VERSION)..."
	@LATEST_BETA=$$(git tag -l "$(VERSION)-beta.*" | sort -V | tail -n 1); \
	if [ -z "$$LATEST_BETA" ]; then \
		echo "❌ No beta versions found for $(VERSION)"; \
		exit 1; \
	fi; \
	echo "   Latest beta: $$LATEST_BETA"; \
	echo ""; \
	echo "🗑️  Removing existing container..."; \
	docker rm -f blockbusterr 2>/dev/null || true; \
	echo ""; \
	echo "🚀 Starting $$LATEST_BETA..."; \
	docker run -d \
		--name blockbusterr \
		-p 9090:9090 \
		-v ~/blockbusterr-data:/app/data \
		ghcr.io/mahcks/blockbusterr:$$LATEST_BETA; \
	if [ $$? -eq 0 ]; then \
		echo ""; \
		echo "✅ Container started successfully!"; \
		echo "   Version: $$LATEST_BETA"; \
		echo "   URL: http://localhost:9090"; \
		echo ""; \
		echo "📝 Useful commands:"; \
		echo "   Logs: docker logs -f blockbusterr"; \
		echo "   Stop: docker stop blockbusterr"; \
	else \
		echo ""; \
		echo "❌ Failed to start container!"; \
		exit 1; \
	fi

# Fast beta build (amd64 only, much faster for testing)
# Usage: make beta-fast VERSION=v1.2.0
beta-fast:
	@if [ -z "$(VERSION)" ]; then \
		echo "Usage: make beta-fast VERSION=v1.2.0"; \
		echo "Example: make beta-fast VERSION=v1.2.0"; \
		exit 1; \
	fi
	@echo "🔄 Fetching latest tags from GitHub..."
	@git fetch --tags --quiet 2>/dev/null || true
	@echo "🔍 Finding latest beta version for $(VERSION)..."
	@LATEST_BETA=$$(git tag -l "$(VERSION)-beta.*" | sort -V | tail -n 1); \
	if [ -z "$$LATEST_BETA" ]; then \
		NEXT_BETA="$(VERSION)-beta.1"; \
		echo "   No existing beta tags found. Starting at $$NEXT_BETA"; \
	else \
		BETA_NUM=$$(echo $$LATEST_BETA | sed 's/.*-beta\.\([0-9]*\)/\1/'); \
		NEXT_NUM=$$((BETA_NUM + 1)); \
		NEXT_BETA="$(VERSION)-beta.$$NEXT_NUM"; \
		echo "   Latest beta: $$LATEST_BETA"; \
		echo "   Next beta: $$NEXT_BETA"; \
	fi; \
	echo ""; \
	echo "📦 Building Docker image (amd64 only - fast build)..."; \
	COMMIT=$$(git rev-parse HEAD); \
	docker buildx build \
		--build-arg VERSION=$$NEXT_BETA \
		--build-arg COMMIT=$$COMMIT \
		--platform linux/amd64 \
		-t ghcr.io/mahcks/blockbusterr:$$NEXT_BETA \
		-t ghcr.io/mahcks/blockbusterr:latest-beta \
		--push \
		.; \
	if [ $$? -ne 0 ]; then \
		echo ""; \
		echo "❌ Build failed!"; \
		exit 1; \
	fi; \
	echo ""; \
	echo "✅ Beta release complete!"; \
	echo "   Version: $$NEXT_BETA"; \
	echo "   Image: ghcr.io/mahcks/blockbusterr:$$NEXT_BETA"; \
	echo "   Platform: linux/amd64 only (fast build)"; \
	echo ""; \
	echo "🏷️  Creating git tag..."; \
	git tag $$NEXT_BETA 2>/dev/null || echo "   Tag already exists locally"; \
	git push origin $$NEXT_BETA 2>/dev/null || echo "   Tag already exists on remote"; \
	echo ""; \
	echo "📝 To test this version:"; \
	echo "   image: ghcr.io/mahcks/blockbusterr:$$NEXT_BETA"

# Full beta build with multi-platform support (slower, for final releases)
# Usage: make beta VERSION=v1.2.0
beta:
	@if [ -z "$(VERSION)" ]; then \
		echo "Usage: make beta VERSION=v1.2.0"; \
		echo "Example: make beta VERSION=v1.2.0"; \
		exit 1; \
	fi
	@echo "🔄 Fetching latest tags from GitHub..."
	@git fetch --tags --quiet 2>/dev/null || true
	@echo "🔍 Finding latest beta version for $(VERSION)..."
	@LATEST_BETA=$$(git tag -l "$(VERSION)-beta.*" | sort -V | tail -n 1); \
	if [ -z "$$LATEST_BETA" ]; then \
		NEXT_BETA="$(VERSION)-beta.1"; \
		echo "   No existing beta tags found. Starting at $$NEXT_BETA"; \
	else \
		BETA_NUM=$$(echo $$LATEST_BETA | sed 's/.*-beta\.\([0-9]*\)/\1/'); \
		NEXT_NUM=$$((BETA_NUM + 1)); \
		NEXT_BETA="$(VERSION)-beta.$$NEXT_NUM"; \
		echo "   Latest beta: $$LATEST_BETA"; \
		echo "   Next beta: $$NEXT_BETA"; \
	fi; \
	echo ""; \
	echo "📦 Building Docker image (multi-platform - this will take a while)..."; \
	COMMIT=$$(git rev-parse HEAD); \
	docker buildx build \
		--build-arg VERSION=$$NEXT_BETA \
		--build-arg COMMIT=$$COMMIT \
		--platform linux/amd64,linux/arm64 \
		-t ghcr.io/mahcks/blockbusterr:$$NEXT_BETA \
		-t ghcr.io/mahcks/blockbusterr:latest-beta \
		--push \
		.; \
	if [ $$? -ne 0 ]; then \
		echo ""; \
		echo "❌ Build failed!"; \
		exit 1; \
	fi; \
	echo ""; \
	echo "✅ Beta release complete!"; \
	echo "   Version: $$NEXT_BETA"; \
	echo "   Image: ghcr.io/mahcks/blockbusterr:$$NEXT_BETA"; \
	echo "   Platforms: linux/amd64, linux/arm64"; \
	echo ""; \
	echo "🏷️  Creating git tag..."; \
	git tag $$NEXT_BETA 2>/dev/null || echo "   Tag already exists locally"; \
	git push origin $$NEXT_BETA 2>/dev/null || echo "   Tag already exists on remote"; \
	echo ""; \
	echo "📝 To test this version:"; \
	echo "   image: ghcr.io/mahcks/blockbusterr:$$NEXT_BETA"

# Development workflow
dev-setup: install build
	@echo "✓ Development environment ready"
	@echo "  Run 'make run' to start the application"

# Quick rebuild and restart (useful during development)
quick: build restart
