BUILD_DIR := "build"
CRAWLER_DIR := "./cmd/crawler"

.PHONY: dep all build clean test lint tidy update_dependencies docker swagger-create swagger-fmt push-tag


all: crawler
build: crawler

lint: # Lint the files
	golangci-lint run

test: ## Run unittests, '-p 1' makes tests not run in parallel;
	go test -p 1 -cover -race ./...

vet: ## Run vet
	go vet ./...

update_dependencies:
	@go get -u ./...

crawler: ## Build the crawler binary
	go build -v -o $(BUILD_DIR)/crawler $(CRAWLER_DIR)

docker: ## Build Docker image for the crawler binary
	docker build -t dakar .

tidy: ## Run go mod tidy on the default go.mod file
	@go mod tidy

clean: ## Remove previous build
	@rm -rf $(BUILD_DIR)

swagger-create: ## Creates openapi docs
	mkdir -p openapi
	./swag init --pd -d cmd/crawler,server -o openapi
ifndef CI
	(cd ../docker && sudo docker compose -f docker-compose-openapi.yml up && sudo docker compose -f docker-compose-openapi.yml rm -fsv)
	(cd openapi/client/typescript-fetch && yarn npm publish)
endif

swagger-fmt: ## Formats swagger annotations
	./swag fmt -d cmd/crawler,server

push-tag: ## increment the most recent tag and push it and any local commits
	bash createTag.sh

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
