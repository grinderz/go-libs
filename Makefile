MAKEFLAGS += --warn-undefined-variables

SHELL := /usr/bin/env bash -o errtrace -o pipefail -o noclobber -o errexit -o nounset

GOLANGCI_LINT_VERSION := 1.64.8
GOLANGCI_LINT_TIMEOUT := 5m

ARGS ?=

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk command is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

DEFAULT_GOAL := help
.PHONY: help
help: ## Display this help screen
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9\-\\.%]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: generate
generate: ## Go generate recursive
	go generate ./...

.PHONY: format
format: ## Format the source code
	# protogetter --fix ./...
	go run github.com/segmentio/golines@latest --max-len=120 --no-reformat-tags --ignore-generated --write-output .
	go run mvdan.cc/gofumpt@latest -l -w -modpath . .
	go run golang.org/x/tools/cmd/goimports@latest -l -w .
	go run github.com/daixiang0/gci@latest write --skip-generated -s standard -s default .

##@ Lint

.PHONY: lint.docker.golangci
lint.docker.golangci: ## Run golangci-lint in docker
	docker run -t --rm -v $$(pwd):/app -v ~/.cache/golangci-lint/v$(GOLANGCI_LINT_VERSION):/root/.cache -w /app golangci/golangci-lint:v$(GOLANGCI_LINT_VERSION) golangci-lint run --timeout=$(GOLANGCI_LINT_TIMEOUT)

.PHONY: lint.golangci
lint.golangci: ## Run golangci-lint
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@v$(GOLANGCI_LINT_VERSION) run --timeout=$(GOLANGCI_LINT_TIMEOUT) $(ARGS)

.PHONY: lint.pre-commit
lint.pre-commit: ## Run pre-commit
	pre-commit run --all-files

.PHONY: lint
lint: lint.golangci lint.pre-commit ## Run all linters

##@ Tests

.PHONY: test
test: ## Run go test
	go test -v -covermode=count -coverprofile=coverage.out -tags coverage ./...

.PHONY: cover
cover: test ## Run coverage report
	go tool cover -func=coverage.out
