BUILD_DIR := build
DAKAR_DIR := ./cmd/dakar

GREEN  = $(shell tput -Txterm setaf 2)
YELLOW = $(shell tput -Txterm setaf 3)
WHITE  = $(shell tput -Txterm setaf 7)
CYAN   = $(shell tput -Txterm setaf 6)
RESET  = $(shell tput -Txterm sgr0)

.PHONY: all build clean test vet lint tidy update_dependencies docker
.PHONY: openapi-fmt openapi-spec openapi-client openapi-publish push-tag spellcheck-wikiapi
.PHONY: licenses

all: dakar
build: dakar

## Test:
lint: ## Lint the files
	cd backend && golangci-lint run

test: ## Run tests
	cd backend && go test -cover -race ./...

vet: ## Run vet
	cd backend && go vet ./...

## Dependencies:
update_dependencies: ## Update Golang dependencies
	cd backend && go get -u ./...

tidy: ## Run go mod tidy on the default go.mod file
	cd backend && go mod tidy

## Build:
dakar: ## Build the dakar binary
	cd backend && go build -v -o $(BUILD_DIR)/dakar $(DAKAR_DIR)

clean: ## Remove previous build
	@rm -rf backend/$(BUILD_DIR)

## Docker:
docker: ## Build Docker image for the dakar binary
	docker build -t dakar backend

## Open API:
openapi-fmt: ## Formats swagger annotations
	cd backend && ./swag fmt -d cmd/dakar,server

openapi-spec: ## Creates openapi spec
	cd backend && mkdir -p openapi && ./swag init --pd -d cmd/dakar,server -o openapi

openapi-client: ## Creates a javascript client based on the openapi spec
	(cd docker && sudo docker compose -f docker-compose-openapi.yml up && sudo docker compose -f docker-compose-openapi.yml rm -fsv)

openapi-publish: ## Publishes an existing client to the configured repository
	(cd backend/openapi/client/typescript-fetch && pnpm publish)

## Git:
push-tag: ## increment the most recent tag and push it and any local commits
	bash backend/createTag.sh

## Spellcheck:
spellcheck-wiki: ## check the spelling of all markdown files in docker/wikiapi/files
	find docker/wikiapi/files/ -type f -name "*.md" -exec aspell --dont-backup -c "{}" \;

## Licenses
licenses: ## generate a JSON file that contains license information of all Javascript and Golang dependencies
	{ cd backend && go-licenses report ./cmd/dakar | grep -v "Unknown" | jq -R '[inputs | split(",") | {name: .[0], license: .[2]}]'; (cd ../app && pnpm licenses ls --prod --json | jq 'map_values(map({name,license})) | [.[] | .[]]'); } | jq -s '{ backend: .[0], frontend: .[1] }' > app/public/licenses.json

## Help:
help: ## Show this help.
	@echo ''
	@echo 'Usage:'
	@echo '  ${YELLOW}make${RESET} ${GREEN}<target>${RESET}'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} { \
		if (/^[a-zA-Z_-]+:.*?##.*$$/) {printf "    ${YELLOW}%-20s${GREEN}%s${RESET}\n", $$1, $$2} \
		else if (/^## .*$$/) {printf "  ${CYAN}%s${RESET}\n", substr($$1,4)} \
		}' $(MAKEFILE_LIST)