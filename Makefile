.PHONY: help build install clean fmt test collect lint

BIN_DIR := bin
BINARY := $(BIN_DIR)/archlint
TRACELINT := $(BIN_DIR)/tracelint
GRAPH_DIR := arch
OUTPUT_ARCH := architecture.yaml

help: ## Show help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Build project
	@mkdir -p $(BIN_DIR)
	go build -o $(BINARY) ./cmd/archlint
	@echo "Build complete: $(BINARY)"

install: build ## Install to GOPATH
	go install ./cmd/archlint

fmt: ## Format code
	go fmt ./...

test: ## Run tests
	go test -v ./...

clean: ## Clean artifacts
	rm -rf $(BIN_DIR)

collect: build ## Build structural graph
	@mkdir -p $(GRAPH_DIR)
	$(BINARY) collect . -o $(GRAPH_DIR)/$(OUTPUT_ARCH)

lint: ## Run tracelint
	@mkdir -p $(BIN_DIR)
	go build -o $(TRACELINT) ./cmd/tracelint
	@$(TRACELINT) ./internal/... ./cmd/archlint/... || true

.DEFAULT_GOAL := help
