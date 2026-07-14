MAKEFLAGS += --warn-undefined-variables

TEST_COVERAGE_THRESHOLD ?= 36
PRE_COMMIT_VERSION := 4.6
PRE_COMMIT_CMD ?= uvx pre-commit@$(PRE_COMMIT_VERSION)
PRE_COMMIT_CFG ?= .pre-commit-config.yaml
GOLANGCI_LINT_VERSION := 2.12.2
GOTESTSUM_VERSION := 1.13.0
GOLANGCI_LINT_TIMEOUT := 5m
SHELL := /usr/bin/env bash -o errtrace -o pipefail -o noclobber -o errexit -o nounset
ARTIFACTS_DIR := artifacts
GOTESTSUB_ARGS ?= --format testname
TEST_CMD := go run gotest.tools/gotestsum@v$(GOTESTSUM_VERSION) $(GOTESTSUB_ARGS) --
TEST_COVERAGE_EXCLUDE ?= (_mock\.go|\/testmocks\/|\.pb\.go)
GO_MOD_ID = github.com/grinderz/go-libs

SHELLCHECK_DIR_EXCLUDE ?= \
	-path ./.git \
	-o -path ./.uv-cache \
	-o -path ./.go

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

.DEFAULT_GOAL := help
.PHONY: help
help: ## Display this help screen
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} \
		/^[a-zA-Z_0-9\-\\.%]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 } \
		/^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

$(ARTIFACTS_DIR):
	mkdir -p $@

.PHONY: clean
clean: ## Cleanup
	@rm -rf "$(ARTIFACTS_DIR)"


##@ Development


.PHONY: go.generate
go.generate: ## Go generate recursive
	go generate ./...

.PHONY: go.format
go.format: ## Format the source code
	# protogetter --fix ./...
	go run github.com/segmentio/golines@latest --max-len=120 --no-reformat-tags --ignore-generated --write-output .
	go run mvdan.cc/gofumpt@latest -l -w -modpath . .
	go run golang.org/x/tools/cmd/goimports@latest -l -w .
	go run github.com/daixiang0/gci@latest write --skip-generated -s standard -s default .


##@ Lint


.PHONY: lint.docker.golangci
lint.docker.golangci: ## Run golangci-lint in docker
	docker run -t --rm \
		-v $$(pwd):/app \
		-v ~/.cache/golangci-lint/v$(GOLANGCI_LINT_VERSION):/root/.cache \
		-w /app \
		golangci/golangci-lint:v$(GOLANGCI_LINT_VERSION) \
		make lint.golangci.bin

.PHONY: lint.golangci
lint.golangci: ## Run golangci-lint
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v$(GOLANGCI_LINT_VERSION) \
		run --timeout=$(GOLANGCI_LINT_TIMEOUT) --show-stats $(ARGS)

.PHONY: lint.golangci.bin
lint.golangci.bin: ## Run golangci-lint bin
	golangci-lint run --timeout=$(GOLANGCI_LINT_TIMEOUT) --show-stats $(ARGS)

.PHONY: lint.shellcheck
lint.shellcheck: ## Run shellcheck
	find . -type d \( $(SHELLCHECK_DIR_EXCLUDE) \) -prune \
		-o -type f -name '*.sh' -exec shellcheck --format=gcc -s bash {} +

.PHONY: lint.pre-commit
lint.pre-commit: ## Run pre-commit
	$(PRE_COMMIT_CMD) run --all-files --config $(PRE_COMMIT_CFG) $(ARGS)

lint.pre-commit.%: ## Run pre-commit with custom command
	$(PRE_COMMIT_CMD) $* --config $(PRE_COMMIT_CFG) $(ARGS)

.PHONY: lint.gitattributes
lint.gitattributes: ## Run gitattributes check
	scripts/gitattributes-check.sh

.PHONY: lint
lint: lint.golangci lint.shellcheck lint.gitattributes lint.pre-commit ## Run all linters


##@ Tests


.PHONY: test
test: ## Run tests
	 $(TEST_CMD) -v -race $(ARGS) $(GO_MOD_ID)/...

.PHONY: test.coverage
test.coverage: ARGS := -tags=coverage -coverpkg=$(GO_MOD_ID)/... -covermode=atomic \
	-coverprofile=$(ARTIFACTS_DIR)/coverage.out.tmp $(ARGS)
test.coverage: $(ARTIFACTS_DIR) test ## Run tests with coverage report
	grep -vE "$(TEST_COVERAGE_EXCLUDE)" $(ARTIFACTS_DIR)/coverage.out.tmp >| $(ARTIFACTS_DIR)/coverage.out
	go tool cover -html=$(ARTIFACTS_DIR)/coverage.out -o $(ARTIFACTS_DIR)/coverage.html
	go tool cover -func=$(ARTIFACTS_DIR)/coverage.out
	./scripts/check-coverage.sh $(ARTIFACTS_DIR)/coverage.out $(TEST_COVERAGE_THRESHOLD)


##@ Deps


.PHONY: deps.update.all.patch
deps.update.all.patch: ## Update all deps to latest patch
	go get -u=patch ./...
	go mod tidy

.PHONY: deps.update.all.latest
deps.update.all.latest: ## Update all deps to latest
	go get -u ./...
	go mod tidy

.PHONY: deps.install.tools
deps.install.tools: ## Install tools from tools.go
	grep _ tools.go | awk -F'"' '{print $$2}' | xargs -tI % go install %
