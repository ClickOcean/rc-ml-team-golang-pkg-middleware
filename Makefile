BINDIR      := $(CURDIR)/bin
BINNAME     ?= app
IS_CI       ?= false

GOLANGCI_VERSION	:= 1.50.1

OS_ID := $(shell (. /etc/os-release; echo $$ID))
ifeq ($(OS_ID),alpine)
    MUSL_TAG = -tags musl
endif
ifeq ($(OS_ID),ubuntu)
    MUSL_TAG =  # empty string, for Ubuntu isn't used
endif

GO := $(shell which go)
ifeq ($(GO),)
    GO = go-is-required
endif

GOPATH := $(shell go env GOPATH)

$(GO):
	@echo "'Golang' is required! Please install it from here: https://go.dev/doc/install." >&2
	@exit 1

GO-LINT := $(shell which golangci-lint)
ifeq ($(GO-LINT),)
    GO-LINT = golangci-lint-is-required
endif

.PHONY: help
help: ## Display this help screen
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: init
init: ## add missing and remove unused modules and downloads modules from go.mod
	go mod tidy

.PHONY: build
build: | init ## build binary
	go build $(MUSL_TAG) -o '$(BINDIR)'/$(BINNAME)

.PHONY: test-coverage
test-coverage: | init ## run tests with coverage
	CGO_ENABLED=0 go run gotest.tools/gotestsum@latest --junitfile report.xml --format testname
	go test ./... -coverprofile=coverage.txt -covermode count
	go get github.com/boumenot/gocover-cobertura
	go run github.com/boumenot/gocover-cobertura < coverage.txt > coverage.xml

.PHONY: test-style
test-style: | $(GO-LINT) ## run linters
ifeq ($(IS_CI),true)
	golangci-lint run --out-format checkstyle --issues-exit-code 0 > reports/lint.xml
else
	golangci-lint run
endif