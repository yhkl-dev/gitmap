.PHONY: build run test clean install lint fmt dev

BINARY := gitmap
CMD := ./cmd/gitmap
GO := go
GOFLAGS := -ldflags="-s -w"

build: ## Build the binary
	$(GO) build $(GOFLAGS) -o $(BINARY) $(CMD)

run: build ## Build and run
	./$(BINARY)

test: ## Run all tests
	$(GO) test ./... -v -count=1

test-race: ## Run tests with race detector
	$(GO) test ./... -race -count=1

cover: ## Run tests with coverage
	$(GO) test ./... -coverprofile=coverage.out
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "coverage report → coverage.html"

clean: ## Remove build artifacts
	rm -f $(BINARY) coverage.out coverage.html

install: ## Install to $GOPATH/bin
	$(GO) install $(CMD)

lint: ## Run static analysis
	$(GO) vet ./...

fmt: ## Format code
	$(GO) fmt ./...

tidy: ## Tidy module dependencies
	$(GO) mod tidy

dev: ## Build, run tests, then run
	$(GO) build $(GOFLAGS) -o $(BINARY) $(CMD) && $(GO) test ./... && ./$(BINARY)

release: ## Run goreleaser to create a release (tag must exist)
	goreleaser release --clean

release-snapshot: ## Run goreleaser in snapshot mode (no publish)
	goreleaser release --clean --snapshot

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*##' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'
