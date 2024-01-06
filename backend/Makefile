BUILD_DIR := "build"
CRAWLER_DIR := "./cmd/crawler"

.PHONY: all dep build dash bitcoin clean test lint tidy update_dependencies docker-dash docker-bitcoin swagger-create swagger-fmt push-tag

all: build

lint: # Lint the files
	golangci-lint run

test: ## Run unittests, '-p 1' makes tests not run in parallel;
	go test -p 1 -cover -race ./...

vet: ## Run vet
	go vet ./...

build: dash bitcoin ## Builds both Dash and Bitcoin

update_dependencies:
	@go get -u ./...
	@go get -u -modfile $(CRAWLER_DIR)/build/bitcoin_go.mod ./...

dash: ## Build the Dash binary
	go build -v -ldflags "-X main.blockchainMode=Dash" -o $(BUILD_DIR)/crawler_dash $(CRAWLER_DIR)

bitcoin: ## Build the Bitcoin binary
	go build -v -ldflags "-X main.blockchainMode=Bitcoin" -modfile $(CRAWLER_DIR)/build/bitcoin_go.mod -o $(BUILD_DIR)/crawler_bitcoin $(CRAWLER_DIR)

docker-dash: ## Build Docker image for the Dash binary
	docker build -t dakar-dash --build-arg="BLOCKCHAIN_MODE=Dash" .

docker-bitcoin: ## Build Docker image for the Bitcoin binary
	docker build -t dakar-bitcoin --build-arg="BLOCKCHAIN_MODE=Bitcoin"\
			--build-arg="MODFILE_PATH=$(CRAWLER_DIR)/build/bitcoin_go.mod"\
			 --build-arg="MODSUM_PATH=$(CRAWLER_DIR)/build/bitcoin_go.sum" .

tidy: ## Run go mod tidy on the default go.mod file and bitcoin go.mod file
	@go mod tidy
	@go mod tidy -modfile ./cmd/crawler/build/bitcoin_go.mod

clean: ## Remove previous build
	@rm -rf $(BUILD_DIR)

swagger-create: ## Creates openapi docs
	./swag init --pd -d cmd/crawler,server -o openapi
ifndef CI
	(cd ../docker && sudo docker-compose -f docker-compose-openapi.yml up && sudo docker-compose -f docker-compose-openapi.yml rm -fsv)
	(cd openapi/client/typescript-fetch && yarn npm publish)
endif

swagger-fmt: ## Formats swagger annotations
	./swag fmt -d cmd/crawler,server

push-tag: ## increment the most recent tag and push it and any local commits
	bash createTag.sh

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
