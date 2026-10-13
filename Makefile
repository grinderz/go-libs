MAKEFLAGS += --warn-undefined-variables

.ONESHELL:
.SHELLFLAGS := -o errtrace -o pipefail -o noclobber -o errexit -o nounset -c
SHELL := bash

# The targets and variables keep the names of a shared make include used by
# other Go projects, inlined here because this module is standalone; scripts/
# carries copies of its scripts. Parts tied to those projects (private
# registry, netrc, compose, env docs, internal deps) are left out.

COMMON_ARTIFACTS_DIR ?= artifacts

COMMON_SHELLCHECK_DIR_EXCLUDE ?= \
	-path ./.git \
	-o -path ./.uv-cache \
	-o -path ./.go

COMMON_PRE_COMMIT_VERSION ?= 4.6
COMMON_PRE_COMMIT_CMD ?= uvx pre-commit@$(COMMON_PRE_COMMIT_VERSION)
COMMON_PRE_COMMIT_GC_CMD ?= $(COMMON_PRE_COMMIT_CMD)
COMMON_PRE_COMMIT_CFG ?= .pre-commit-config.yaml

# recursive: only the docker targets expand it, so an unsupported host
# architecture fails those alone instead of every make run
COMMON_ARCH = $(shell scripts/architecture-detection.sh)

# make target inside a toolchain image with the sources bind mounted. The
# container is root, whatever it wrote into the tree is chowned back on exit
COMMON_DOCKER_APP_DIR ?= /app
COMMON_DOCKER_CHOWN_CMD ?= find $(COMMON_DOCKER_APP_DIR) -xdev ! -user $(shell id -u) -exec chown -h $(shell id -u):$(shell id -g) {} + 2>/dev/null

# $(1): image, $(2): docker run args, $(3): command before make, $(4): target.
# Make variables like ARGS pass as env through -e in $(2), a command line
# override would replace the target-specific values of the coverage targets.
# make runs in the background: sh defers a trap while a foreground command
# runs, wait picks a ctrl-c up at once and the trap forwards it to everything
# --platform: a tag pulled for another platform, like the arm64 half of an
# image checked on an amd64 host, otherwise stays what docker run picks
define common-docker-run
	docker run -t --rm --init --platform linux/$(COMMON_ARCH) \
		$(2) \
		-v "$$(pwd)":$(COMMON_DOCKER_APP_DIR) \
		-w $(COMMON_DOCKER_APP_DIR) \
		$(1) \
		sh -c 'trap "$(COMMON_DOCKER_CHOWN_CMD) || :" EXIT; \
			$(3) || exit; \
			trap "kill -s INT -- -1" INT; trap "kill -s TERM -- -1" TERM; \
			make $(4) & wait $$!; st=$$?; wait; exit $$st'
endef

# golanglint
GO_GOLANGCI_LINT_VERSION ?= 2.14.0
GO_GOLANGCI_LINT_TIMEOUT ?= 5m

# the staticcheck bundled with golangci-lint supports the current go release
# and fails on a newer one, so a local run uses the toolchain of go.mod's
# toolchain line. That line is only a minimum, a newer local toolchain still
# wins without this. In ci setup-go already installs the right toolchain.
GO_TOOLCHAIN ?= go1.26.9

ifndef CI
export GOTOOLCHAIN = $(GO_TOOLCHAIN)
endif

# test
GO_TEST_COVERAGE_THRESHOLD ?= 48
GO_GOTESTSUM_VERSION ?= 1.13.0
GO_GOTESTSUM_ARGS ?= --format testname
GO_GOTESTSUM_CMD ?= go run gotest.tools/gotestsum@v$(GO_GOTESTSUM_VERSION) $(GO_GOTESTSUM_ARGS) --
GO_TEST_CMD ?= $(GO_GOTESTSUM_CMD)
GO_TEST_COVERAGE_EXCLUDE ?= (_mock\.go|/testmocks/|\.pb\.go)
GO_TEST_REPEAT_COUNT ?= 2
# on: a random order per pass, the seed is printed on failure and can be
# passed back here to reproduce it
GO_TEST_SHUFFLE ?= on
# build tags for the tests, space or comma separated. A repeated -tags flag
# replaces the earlier one, so all tags are joined into a single flag here and
# test.coverage adds its own tag to the list instead of passing a second flag
GO_TEST_TAGS ?=
go-empty :=
go-space := $(go-empty) $(go-empty)
go-comma := ,
GO_TEST_TAGS_FLAG = $(if $(strip $(GO_TEST_TAGS)),-tags=$(subst $(go-space),$(go-comma),$(strip $(subst $(go-comma),$(go-space),$(GO_TEST_TAGS)))))

# debian: a glibc run, alpine: a musl/cgo run. The exact patch of
# GO_TOOLCHAIN keeps the image and the pin above in sync, so go inside the
# container never downloads another toolchain
GO_DOCKER_GO_VERSION ?= $(patsubst go%,%,$(GO_TOOLCHAIN))
GO_DOCKER_DEBIAN_VERSION ?= trixie
GO_DOCKER_ALPINE_VERSION ?= 3.24
GO_DOCKER_IMAGE ?= golang:$(GO_DOCKER_GO_VERSION)-$(GO_DOCKER_DEBIAN_VERSION)
GO_DOCKER_ALPINE_IMAGE ?= golang:$(GO_DOCKER_GO_VERSION)-alpine$(GO_DOCKER_ALPINE_VERSION)
GO_DOCKER_ALPINE_SETUP_CMD ?= apk add --no-cache bash make gcc musl-dev git
GO_DOCKER_ALPINE_TAGS ?= musl nomsgpack
# runs before make inside the container, the alpine image lacks make and gcc
GO_DOCKER_SETUP_CMD ?= :
# named volumes: without them every run downloads the modules and rebuilds all
GO_DOCKER_MOD_CACHE_VOLUME ?= go-libs-go-mod-cache
GO_DOCKER_BUILD_CACHE_VOLUME ?= go-libs-go-build-cache
GO_DOCKER_RUN_ARGS ?= \
	-e ARGS -e GO_TEST_TAGS \
	-v $(GO_DOCKER_MOD_CACHE_VOLUME):/go/pkg/mod \
	-v $(GO_DOCKER_BUILD_CACHE_VOLUME):/root/.cache/go-build

GO_MOD_ID ?= github.com/grinderz/go-libs

ARGS ?=


##@ Common General


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

$(COMMON_ARTIFACTS_DIR):
	mkdir -p $@

.PHONY: clean
clean: ## Cleanup
	rm -rf "$(COMMON_ARTIFACTS_DIR)"


##@ Common Lint


.PHONY: lint.shellcheck
lint.shellcheck: ## Run shellcheck
	find . -type d \( $(COMMON_SHELLCHECK_DIR_EXCLUDE) \) -prune \
		-o -type f -name '*.sh' -exec shellcheck --format=gcc -s bash {} +

.PHONY: lint.pre-commit
lint.pre-commit: ## Run pre-commit
	$(COMMON_PRE_COMMIT_CMD) run --all-files --config $(COMMON_PRE_COMMIT_CFG) $(ARGS)

lint.pre-commit.%: ## Run pre-commit with custom command
	$(COMMON_PRE_COMMIT_CMD) $* --config $(COMMON_PRE_COMMIT_CFG) $(ARGS)

# gc accepts no --config option, so it cannot go through the pattern rule above
.PHONY: lint.pre-commit.gc
lint.pre-commit.gc: ## Run pre-commit gc
	$(COMMON_PRE_COMMIT_GC_CMD) gc $(ARGS)

.PHONY: lint.gitattributes
lint.gitattributes: ## Run gitattributes check
	scripts/gitattributes-check.sh


##@ Go Development


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


##@ Go Lint


.PHONY: lint.docker.golangci
lint.docker.golangci: ## Run golangci-lint in docker
	docker run -t --rm \
		-v $$(pwd):/app \
		-v ~/.cache/golangci-lint/v$(GO_GOLANGCI_LINT_VERSION):/root/.cache \
		-w /app \
		golangci/golangci-lint:v$(GO_GOLANGCI_LINT_VERSION) \
		make lint.golangci.bin

.PHONY: lint.golangci
lint.golangci: ## Run golangci-lint
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v$(GO_GOLANGCI_LINT_VERSION) \
		run --timeout=$(GO_GOLANGCI_LINT_TIMEOUT) --show-stats $(ARGS)

.PHONY: lint.golangci.bin
lint.golangci.bin: ## Run golangci-lint bin
	golangci-lint run --timeout=$(GO_GOLANGCI_LINT_TIMEOUT) --show-stats $(ARGS)

.PHONY: lint.fix
lint.fix: ## Fix code
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v$(GO_GOLANGCI_LINT_VERSION) \
		run --timeout=$(GO_GOLANGCI_LINT_TIMEOUT) --fix --show-stats $(ARGS)

.PHONY: lint
lint: lint.golangci lint.shellcheck lint.gitattributes lint.pre-commit ## Run all linters


##@ Go Test


.PHONY: test
test: ## Run tests
	$(GO_TEST_CMD) -v -race $(GO_TEST_TAGS_FLAG) $(ARGS) $(GO_MOD_ID)/...

# runs every test N times in one process and in a shuffled order: state that
# leaks between runs, like a logger installed twice through a package global,
# only shows up on the second pass, and order dependence only in a different
# order, both stay invisible to a plain make test
.PHONY: test.repeat
test.repeat: override ARGS := -count=$(GO_TEST_REPEAT_COUNT) -shuffle=$(GO_TEST_SHUFFLE) $(ARGS)
test.repeat: test ## Run tests GO_TEST_REPEAT_COUNT times in one process, shuffled

# override: a command line GO_TEST_TAGS= or ARGS= extends these instead of
# replacing them, otherwise the coverage tag or the coverprofile flag is lost
.PHONY: test.coverage
test.coverage: override GO_TEST_TAGS := coverage $(GO_TEST_TAGS)
test.coverage: override ARGS := -coverpkg=$(GO_MOD_ID)/... -covermode=atomic -coverprofile=$(COMMON_ARTIFACTS_DIR)/coverage.out.tmp $(ARGS)
test.coverage: $(COMMON_ARTIFACTS_DIR) test ## Run tests with coverage report
	grep -vE "$(GO_TEST_COVERAGE_EXCLUDE)" $(COMMON_ARTIFACTS_DIR)/coverage.out.tmp >| $(COMMON_ARTIFACTS_DIR)/coverage.out
	go tool cover -html=$(COMMON_ARTIFACTS_DIR)/coverage.out -o $(COMMON_ARTIFACTS_DIR)/coverage.html
	go tool cover -func=$(COMMON_ARTIFACTS_DIR)/coverage.out
	./scripts/check-coverage.sh $(COMMON_ARTIFACTS_DIR)/coverage.out $(GO_TEST_COVERAGE_THRESHOLD)

define go-docker-run
	$(call common-docker-run,$(GO_DOCKER_IMAGE),$(GO_DOCKER_RUN_ARGS),$(GO_DOCKER_SETUP_CMD),$(1))
endef

.PHONY: test.docker
test.docker: ## Run tests in the debian golang image, ARGS= passes go test args
	$(call go-docker-run,test)

test.docker.%: ## Run make test.$* in the debian golang image, e.g. test.docker.coverage
	$(call go-docker-run,test.$*)

.PHONY: test.docker.alpine
test.docker.alpine: ## Run tests in the alpine golang image with the musl build tags
	@$(MAKE) test.docker GO_DOCKER_IMAGE="$(GO_DOCKER_ALPINE_IMAGE)" GO_DOCKER_SETUP_CMD="$(GO_DOCKER_ALPINE_SETUP_CMD)" GO_TEST_TAGS="$(GO_DOCKER_ALPINE_TAGS) $(GO_TEST_TAGS)"

test.docker.alpine.%: ## Run make test.$* in the alpine golang image with the musl build tags, e.g. test.docker.alpine.coverage
	@$(MAKE) test.docker.$* GO_DOCKER_IMAGE="$(GO_DOCKER_ALPINE_IMAGE)" GO_DOCKER_SETUP_CMD="$(GO_DOCKER_ALPINE_SETUP_CMD)" GO_TEST_TAGS="$(GO_DOCKER_ALPINE_TAGS) $(GO_TEST_TAGS)"


##@ Go Deps


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
