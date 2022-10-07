BUILD_DIR := "build"
PKG_LIST := $(shell go list ./... | grep -v /vendor/)

.PHONY: all dep build clean test lint tidy

all: build

lint: # Lint the files
	@go install honnef.co/go/tools/cmd/staticcheck@latest
	@${CI_PROJECT_DIR}/vendor/go/bin/staticcheck -fail ${PKG_LIST}

test: ## Run unittests
	@go test -cover ${PKG_LIST}

build: dash bitcoin ## Builds both Dash and Bitcoin

dash: ## Build the Dash binary
	go build -v -ldflags "-X main.blockchainMode=Dash" -o $(BUILD_DIR)/crawler_dash ./cmd/crawler

bitcoin: ## Build the Bitcoin binary
	go build -v -ldflags "-X main.blockchainMode=Bitcoin" -modfile ./cmd/crawler/build/bitcoin_go.mod -o $(BUILD_DIR)/crawler_bitcoin ./cmd/crawler

tidy: ## Run go mod tidy on the default go.mod file and bitcoin go.mod file
	go mod tidy
	go mod tidy -modfile ./cmd/crawler/build/bitcoin_go.mod

clean: ## Remove previous build
	@rm -rf $(BUILD_DIR)

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
