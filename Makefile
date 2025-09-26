.PHONY: test build build-tools build-optimized install install-tools install-optimized tui dev clean benchmark info

# Go 1.25 optimization flags (enabled by default for performance)
GO_EXPERIMENT := jsonv2,greenteagc
GO_OPTIMIZED := GOEXPERIMENT=$(GO_EXPERIMENT)
BUILD_RELEASE_FLAGS := -ldflags="-w -s" -trimpath

test:
	$(GO_OPTIMIZED) go test ./...

# Build for current platform with Go 1.25 optimizations (DEFAULT)
build:
	@echo "🚀 Building with Go 1.25 optimizations (JSON v2 + Green Tea GC)..."
ifeq ($(OS),Windows_NT)
	$(GO_OPTIMIZED) go build -o agentry.exe ./cmd/agentry
else
	$(GO_OPTIMIZED) go build -o agentry ./cmd/agentry
endif
	@echo "✅ Optimized build complete"

# Standard build without optimizations (for compatibility)
build-standard:
	@echo "📋 Building standard version (no experimental features)..."
ifeq ($(OS),Windows_NT)
	go build -o agentry-standard.exe ./cmd/agentry
else
	go build -o agentry-standard ./cmd/agentry
endif
	@echo "✅ Standard build complete"

# Build with tools tag for full functionality (optimized)
build-tools:
	@echo "🔧 Building with tools and Go 1.25 optimizations..."
ifeq ($(OS),Windows_NT)
	$(GO_OPTIMIZED) go build -tags tools -o agentry.exe ./cmd/agentry
else
	$(GO_OPTIMIZED) go build -tags tools -o agentry ./cmd/agentry
endif
	@echo "✅ Optimized tools build complete"

# Release build (fully optimized)
build-optimized:
	@echo "📦 Building optimized release version..."
ifeq ($(OS),Windows_NT)
	$(GO_OPTIMIZED) go build $(BUILD_RELEASE_FLAGS) -o agentry.exe ./cmd/agentry
else
	$(GO_OPTIMIZED) go build $(BUILD_RELEASE_FLAGS) -o agentry ./cmd/agentry
endif
	@echo "✅ Release build complete"

# Install optimized version to Go's bin directory (DEFAULT)
install:
	@echo "📦 Installing optimized Agentry..."
	$(GO_OPTIMIZED) go install ./cmd/agentry
	@echo "✅ Optimized Agentry installed"

# Install with tools tag (optimized)
install-tools:
	@echo "📦 Installing optimized Agentry with tools..."
	$(GO_OPTIMIZED) go install -tags tools ./cmd/agentry
	@echo "✅ Optimized Agentry with tools installed"

# Install standard version (no optimizations)
install-standard:
	@echo "📦 Installing standard Agentry..."
	go install ./cmd/agentry
	@echo "✅ Standard Agentry installed"

# Run performance benchmark
benchmark:
	@echo "⚡ Running JSON performance benchmark with optimizations..."
	$(GO_OPTIMIZED) go run ./cmd/benchmark 5000

# Show build information
info:
	@echo "Agentry Build Configuration:"
	@echo "  Go Version: $(shell go version)"
	@echo "  Default Optimizations: $(GO_EXPERIMENT)"
	@echo "  Release Flags: $(BUILD_RELEASE_FLAGS)"
	@echo ""
	@echo "Build targets:"
	@echo "  build (default) - Optimized build with Go 1.25 features"
	@echo "  build-standard  - Standard build without experimental features"
	@echo "  build-tools     - Optimized build with tools tag"
	@echo "  build-optimized - Full release optimization"

# Clean up build artifacts (both possible names)
clean:
	@echo "🧹 Cleaning build artifacts..."
	rm -f agentry.exe agentry agentry-standard agentry-standard.exe
	@echo "✅ Clean complete"

# Run optimized TUI
tui: build
ifeq ($(OS),Windows_NT)
	.\agentry.exe tui
else
	./agentry tui
endif

# Development workflow (optimized)
dev: test build
ifeq ($(OS),Windows_NT)
	.\agentry.exe tui
else
	./agentry tui
endif
