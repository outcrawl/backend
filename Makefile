.PHONY: build
build: ## Build Lambda binaries
	GOOS=linux GOARCH=amd64 go build -o build/subscribe ./cmd/subscribe

.PHONY: help
help: ## Display this help screen
	grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
