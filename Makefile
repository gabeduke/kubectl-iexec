.PHONY: help build lint lint-md test snapshot clean
.DEFAULT_GOAL := help

build: ## Build the local kubectl-iexec binary
	go build -ldflags "-s" -o kubectl-iexec .

lint: ## Run golangci-lint (config: .golangci.yml)
	golangci-lint run

# Same markdownlint-cli version reviewdog/action-markdownlint pins. CI only
# reports the lines a PR touches and the tree carries old findings, so pass
# FILES=<the .md files you changed> to see what CI will flag.
lint-md: ## Run markdownlint-cli as CI does (FILES=<paths>, default: whole tree)
	npx --yes markdownlint-cli@0.49.1 --ignore dist $(or $(FILES),.)

test: ## Run tests with the race detector
	go test -race ./...

snapshot: ## Build a goreleaser snapshot (dry-run release)
	goreleaser release --snapshot --clean

clean: ## Remove build artifacts
	rm -f kubectl-iexec
	rm -rf dist/

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
