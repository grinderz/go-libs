MAKEFLAGS += --warn-undefined-variables

TEST_COVERAGE_THRESHOLD ?= 4
PRE_COMMIT_VERSION := 4.5
GOLANGCI_LINT_VERSION := 2.6.2
GOTESTSUM_VERSION := 1.13.0
GOLANGCI_LINT_TIMEOUT := 5m
SHELL := /usr/bin/env bash -o errtrace -o pipefail -o noclobber -o errexit -o nounset
ARTIFACTS_PATH := artifacts
GOTESTSUB_ARGS ?= --format testname
TEST_CMD := go run gotest.tools/gotestsum@v$(GOTESTSUM_VERSION) $(GOTESTSUB_ARGS) --

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

$(ARTIFACTS_PATH):
	mkdir -p $@

.PHONY: clean
clean: ## Cleanup
	@rm -rf "$(ARTIFACTS_PATH)"


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
	docker run -t --rm -v $$(pwd):/app -v ~/.cache/golangci-lint/v$(GOLANGCI_LINT_VERSION):/root/.cache -w /app golangci/golangci-lint:v$(GOLANGCI_LINT_VERSION) make lint.golangci.bin

.PHONY: lint.golangci
lint.golangci: ## Run golangci-lint
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v$(GOLANGCI_LINT_VERSION) run --timeout=$(GOLANGCI_LINT_TIMEOUT) --show-stats $(ARGS)

.PHONY: lint.golangci.bin
lint.golangci.bin: ## Run golangci-lint bin
	golangci-lint run --timeout=$(GOLANGCI_LINT_TIMEOUT) --show-stats

.PHONY: lint.shellcheck
lint.shellcheck: ## Run shellcheck
	find . -type d \( -path ./git -o -path ./.uv-cache -o -path ./.go \) -prune -o -type f -name '*.sh' -exec shellcheck --format=gcc --severity=warning -s bash {}  +

.PHONY: lint.pre-commit
lint.pre-commit: ## Run pre-commit
	uvx pre-commit@$(PRE_COMMIT_VERSION) run --all-files $(ARGS)

.PHONY: lint
lint: lint.golangci lint.shellcheck lint.pre-commit ## Run all linters


##@ Tests


.PHONY: test
test: ## Run tests
	 $(TEST_CMD) -v -race $(ARGS) ./...

.PHONY: test.coverage ## Run test with coverage report
test.coverage: ARGS := -tags=coverage -covermode=atomic -coverprofile=$(ARTIFACTS_PATH)/coverage.out.tmp $(ARGS)
test.coverage: $(ARTIFACTS_PATH) test ## Run tests with coverage
	grep -v "_mock.go" $(ARTIFACTS_PATH)/coverage.out.tmp >| $(ARTIFACTS_PATH)/coverage.out
	go tool cover -html=$(ARTIFACTS_PATH)/coverage.out -o $(ARTIFACTS_PATH)/coverage.html
	go tool cover -func=$(ARTIFACTS_PATH)/coverage.out
	./scripts/check-coverage.sh $(ARTIFACTS_PATH)/coverage.out $(TEST_COVERAGE_THRESHOLD)


.PHONY: test.docker.alpine
test.docker.alpine: GO_VERSION := 1.24
test.docker.alpine: ALPINE_VERSION := 3.21
test.docker.alpine: ## Run test container in alpine image, you can start test by make test ARGS="-tags=musl,nomsgpack"
	docker run -ti --rm \
		-e GOPROXY=https://artifactory.mts.ai/artifactory/go-proxy \
		-e GONOSUMDB=git.mts.ai/ \
		-e GOPRIVATE=git.mts.ai/ \
		-e CGO_ENABLED=1 \
		-v $$(pwd):/app \
		-v ~/.netrc:/root/.netrc:ro  \
		-w /app \
		artifactory.mts.ai/docker-hub-proxy/golang:$(GO_VERSION)-alpine$(ALPINE_VERSION) \
		sh -c "apk update && apk add bash make git musl-dev gcc && /bin/bash"
