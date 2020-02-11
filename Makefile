.PHONY: setup
setup: ## Install all the build and lint dependencies
	go get -u golang.org/x/tools/cmd/cover

.PHONY: test
test: ## Run all the tests
	echo 'mode: atomic' > coverage.txt && go test -covermode=atomic -coverpkg=./... -coverprofile=coverage.txt -race -timeout=30s ./...

.PHONY: cover
cover: test ## Run all the tests and opens the coverage report
	go tool cover -html=coverage.txt

.PHONY: fmt
fmt: ## Run goimports on all go files
	find . -name '*.go' -not -wholename './vendor/*' | while read -r file; do goimports -w "$$file"; done

.PHONY: ci
ci: test ## Run all the tests and code checks

.PHONY: build
build: ## Build a version
	go build -v ./...

.PHONY: clean
clean: ## Remove temporary files
	go clean

# Absolutely awesome: http://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := build
