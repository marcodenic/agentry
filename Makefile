.PHONY: test build install serve dev clean

test:
	go test ./...
	cd ts-sdk && npm install && npm test

# Build to root directory (convenient for development and global usage)
build:
	go build -o agentry.exe ./cmd/agentry

# Install to Go's bin directory (standard Go way)
install:
	go install ./cmd/agentry

# Clean up build artifacts
clean:
	rm -f agentry.exe agentry

serve: build
	.\agentry.exe tui

dev: test build
	.\agentry.exe tui
