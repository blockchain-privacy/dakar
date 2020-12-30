PROJECT_NAME := "dakar"
PKG_LIST := $(shell go list ./... | grep -v /vendor/)

.PHONY: all dep build clean test lint #coverage coverhtml

all: build

lint: # Lint the files
	@go get -u golang.org/x/lint/golint
	@${CI_PROJECT_DIR}/vendor/go/bin/golint -set_exit_status ${PKG_LIST}

test: ## Run unittests
	@go test -short ${PKG_LIST}

race: ## Run data race detector
	@go test -race -short ${PKG_LIST}

# msan: dep ## Run memory sanitizer
# 	@go test -msan -short ${PKG_LIST}

#coverage: ## Generate global code coverage report
#	./tools/coverage.sh;

#coverhtml: ## Generate global code coverage report in HTML
#	./tools/coverage.sh html;

build: ## Build the binary file
	@go build -i -v

clean: ## Remove previous build
	@rm -f $(PROJECT_NAME)

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
