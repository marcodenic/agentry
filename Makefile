.PHONY: test build build-tools install install-tools serve dev clean

test:
	go test ./...

# Build for current platform (Unix/Linux/macOS gets 'agentry', Windows gets 'agentry.exe')
build:
ifeq ($(OS),Windows_NT)
	go build -o agentry.exe ./cmd/agentry
else
	go build -o agentry ./cmd/agentry
endif

# Build with tools tag for full functionality
build-tools:
ifeq ($(OS),Windows_NT)
	go build -tags tools -o agentry.exe ./cmd/agentry
else
	go build -tags tools -o agentry ./cmd/agentry
endif

# Install to Go's bin directory (standard Go way)
install:
	go install ./cmd/agentry

# Install with tools tag
install-tools:
	go install -tags tools ./cmd/agentry

# Clean up build artifacts (both possible names)
clean:
	rm -f agentry.exe agentry

serve: build
ifeq ($(OS),Windows_NT)
	.\agentry.exe tui
else
	./agentry tui
endif

dev: test build
ifeq ($(OS),Windows_NT)
	.\agentry.exe tui
else
	./agentry tui
endif
