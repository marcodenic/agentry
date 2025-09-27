.PHONY: build test install clean run lint info

GO ?= go
BIN_DIR ?= bin
BINARY ?= agentry
GOEXPERIMENT ?= jsonv2,greenteagc
GOENV := GOEXPERIMENT=$(GOEXPERIMENT)

build: ## Build the TUI binary with Go experiments enabled
	@echo "🚀 building agentry ($(GOEXPERIMENT))"
	$(GOENV) $(GO) build -o $(BINARY) ./cmd/agentry
	@echo "✅ output: $(BINARY)"

build-release: ## Build stripped release binary
	@echo "📦 building release binary"
	$(GOENV) $(GO) build -ldflags='-w -s' -trimpath -o $(BINARY) ./cmd/agentry

install: ## Install the binary into GOPATH/bin
	@echo "📦 installing agentry"
	$(GOENV) $(GO) install ./cmd/agentry

test: ## Run the full unit test suite with experiments enabled
	$(GOENV) $(GO) test ./...

run: build ## Build and run the TUI
	$(BINARY) tui

lint: ## Run static checks if golangci-lint is available
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ; \
	else \
		echo "golangci-lint not installed; skipping lint" ; \
	fi

clean: ## Remove build artefacts
	rm -rf $(BIN_DIR)

info: ## Print build configuration
	@echo "Go: $(shell $(GO) version)"
	@echo "GOEXPERIMENT: $(GOEXPERIMENT)"
	@echo "Output: $(BINARY)"
