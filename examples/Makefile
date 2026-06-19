.PHONY: gen test tidy clean help

# Build the generator from the sibling module and put it on PATH for go:generate.
BIN := $(CURDIR)/bin
export PATH := $(BIN):$(PATH)

gen: ## Build the generator and run go:generate to (re)create *_deps.go
	cd ../cmd/depx && go build -o $(BIN)/depx .
	go generate ./...

test: ## Run the generated-constructor tests
	go test -race ./...

tidy: ## Tidy modules
	go mod tidy

clean: ## Remove build artifacts
	rm -rf bin

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-8s\033[0m %s\n", $$1, $$2}'
