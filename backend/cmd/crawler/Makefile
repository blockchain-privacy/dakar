PROJECT_NAME := "dakar"
PKG_LIST := $(shell go list ../../... | grep -v /vendor/)

.PHONY: all dep build clean test lint #coverage coverhtml

all: build

lint: # Lint the files
	@go install honnef.co/go/tools/cmd/staticcheck@latest
	@${CI_PROJECT_DIR}/vendor/go/bin/staticcheck -fail ${PKG_LIST}

test: ## Run unittests
	@go test -short ${PKG_LIST}
	@go test -race -short ${PKG_LIST}

vet: ## Run vet
	@go vet ${PKG_LIST}

build: dash ## Builds Dash by default

dash: ## Build the Dash binary
	go build -v -ldflags "-X main.blockchainMode=Dash"

bitcoin: ## Build the Bitcoin binary
	go build -v -ldflags "-X main.blockchainMode=Bitcoin" -modfile build/go.mod

clean: ## Remove previous build
	@rm -f $(PROJECT_NAME)

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
