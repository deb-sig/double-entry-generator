# Copyright 2019 The Caicloud Authors.
#
# The old school Makefile, following are required targets. The Makefile is written
# to allow building multiple binaries. You are free to add more targets or change
# existing implementations, as long as the semantics are preserved.
# Ref https://github.com/caicloud/golang-template-project/blob/master/Makefile

# This repo's root import path (under GOPATH).
ROOT := github.com/deb-sig/double-entry-generator/v2

# Target binaries. You can build multiple binaries for a single project.
TARGETS := double-entry-generator

# Container image prefix and suffix added to targets.
# The final built images are:
#   $[REGISTRY]/$[IMAGE_PREFIX]$[TARGET]$[IMAGE_SUFFIX]:$[VERSION]
# $[REGISTRY] is an item from $[REGISTRIES], $[TARGET] is an item from $[TARGETS].
IMAGE_PREFIX ?= $(strip template-)
IMAGE_SUFFIX ?= $(strip )

# It's necessary to set this because some environments don't link sh -> bash.
SHELL := /bin/bash

# Project main package location (can be multiple ones).
CMD_DIR := .

# Project output directory.
OUTPUT_DIR := ./bin

# Build direcotory.
BUILD_DIR := ./build

# Current version of the project.
GIT_COMMIT = $(shell git describe --tags --always --dirty)
VERSION ?= ${GIT_COMMIT}

# Track code version with Docker Label.
DOCKER_LABELS ?= git-describe="$(shell date -u +v%Y%m%d)-$(shell git describe --tags --always --dirty)"

# Golang standard bin directory.
GOPATH ?= $(shell go env GOPATH)
BIN_DIR := $(GOPATH)/bin
GOLANGCI_LINT := $(BIN_DIR)/golangci-lint

# https://stackoverflow.com/a/14777895
ifeq ($(OS),Windows_NT)     # is Windows_NT on XP, 2000, 7, Vista, 10...
    detected_OS := Windows
else
    detected_OS := $(shell uname)  # same as "uname -s"
endif

ifeq ($(detected_OS), Windows)
    BROWSER := start
else 
	ifeq (,$(findstring Drawin,$(detected_OS)))  # Mac OS X
    	BROWSER := open
	else  # consider as Linux or others
		BROWSER := xdg-open
	endif
endif

# All targets.
.PHONY: lint test build container push help clean test-go test-wechat test-alipay test-huobi test-htsec format check-format goreleaser-build-test install-golangci-lint clean-cache gen-doc before-commit-check

help:  ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z0-9_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

build: build-local  ## Build the project

LD_FLAGS := -ldflags "-s -w
LD_FLAGS += -X $(ROOT)/pkg/version.VERSION=$(VERSION)
LD_FLAGS += -X $(ROOT)/pkg/version.REPOROOT=$(ROOT)
LD_FLAGS += -X $(ROOT)/pkg/version.COMMIT=$(GIT_COMMIT)"

build-local:
	@for target in $(TARGETS); do                                                      \
	  go build -v -o $(OUTPUT_DIR)/$${target}                                       \
	  $(LD_FLAGS)                        \
	  $(CMD_DIR)/;                                                                     \
	done

install: build  ## Install the double-entry-generator binary
	@install ./bin/double-entry-generator /usr/local/bin
	@echo "Installed double-entry-generator at /usr/local/bin/double-entry-generator !"

clean: ## Clean all the temporary files
	@rm -rf ./bin
	@rm -rf ./dist
	@rm -rf ./test/output
	@rm -rf ./double-entry-generator
	@rm -rf ./wasm-dist

test: test-go test-alipay-beancount test-alipay-ledger test-wechat-beancount test-wechat-ledger test-huobi-beancount test-huobi-ledger test-htsec-beancount test-htsec-ledger test-icbc-beancount test-icbc-ledger test-ccb-beancount test-ccb-ledger test-td-beancount test-td-ledger test-bmo-beancount test-bmo-ledger test-bocomcredit-beancount test-bocomcredit-ledger test-citic-beancount test-citic-ledger test-hsbchk-beancount test-hsbchk-ledger test-jd-beancount test-jd-ledger test-mt-beancount test-hxsec-beancount test-cmb-beancount test-cmb-ledger ## Run all tests

test-go: ## Run Golang tests
	@go test ./...

test-alipay-beancount: ## Run tests for Alipay provider against beancount compiler
	@$(SHELL) ./test/alipay-test-beancount.sh
test-alipay-ledger: ## Run tests for Alipay provider against ledger compiler
	@$(SHELL) ./test/alipay-test-ledger.sh

test-wechat-beancount: ## Run tests for WeChat provider against beancount compiler
	@$(SHELL) ./test/wechat-test-beancount.sh
test-wechat-ledger: ## Run tests for WeChat provider against ledger compiler
	@$(SHELL) ./test/wechat-test-ledger.sh

test-huobi-beancount: ## Run tests for huobi provider against beancount compiler
	@$(SHELL) ./test/huobi-test-beancount.sh
test-huobi-ledger: ## Run tests for huobi provider against ledger compiler
	@$(SHELL) ./test/huobi-test-ledger.sh

test-htsec-beancount: ## Run tests for htsec provider against beancount compiler
	@$(SHELL) ./test/htsec-test-beancount.sh
test-htsec-ledger: ## Run tests for htsec provider against ledger compiler
	@$(SHELL) ./test/htsec-test-ledger.sh

test-icbc-beancount: ## Run tests for ICBC provider against beancount compiler
	@$(SHELL) ./test/icbc-test-beancount.sh
test-icbc-ledger: ## Run tests for ICBC provider against ledger compiler
	@$(SHELL) ./test/icbc-test-ledger.sh

test-ccb-beancount: ## Run tests for CCB provider against beancount compiler
	@$(SHELL) ./test/ccb-test-beancount.sh
test-ccb-ledger: ## Run tests for CCB provider against ledger compiler
	@$(SHELL) ./test/ccb-test-ledger.sh

test-td-beancount: ## Run tests for TD provider against beancount compiler
	@$(SHELL) ./test/td-test-beancount.sh
test-td-ledger: ## Run tests for TD provider against ledger compiler
	@$(SHELL) ./test/td-test-ledger.sh

test-bmo-beancount: ## Run tests for BMO provider against beancount compiler
	@$(SHELL) ./test/bmo-test-beancount.sh
test-bmo-ledger: ## Run tests for BMO provider against ledger compiler
	@$(SHELL) ./test/bmo-test-ledger.sh

test-bocomcredit-beancount: ## Run tests for Bocom Credit provider against beancount compiler
	@$(SHELL) ./test/bocomcredit-test-beancount.sh
test-bocomcredit-ledger: ## Run tests for Bocom Credit provider against ledger compiler
	@$(SHELL) ./test/bocomcredit-test-ledger.sh

test-citic-beancount: ## Run tests for CITIC provider against beancount compiler
	@$(SHELL) ./test/citic-test-beancount.sh
test-citic-ledger: ## Run tests for CITIC provider against ledger compiler
	@$(SHELL) ./test/citic-test-ledger.sh

test-hsbchk-beancount: ## Run tests for HSBC HK provider against beancount compiler
	@$(SHELL) ./test/hsbchk-test-beancount.sh
test-hsbchk-ledger: ## Run tests for HSBC HK provider against ledger compiler
	@$(SHELL) ./test/hsbchk-test-ledger.sh

test-jd-beancount: ## Run tests for jd provider against beancount compiler
	@$(SHELL) ./test/jd-test-beancount.sh
test-jd-ledger: ## Run tests for jd provider against ledger compiler
	@$(SHELL) ./test/jd-test-ledger.sh

test-mt-beancount: ## Run tests for mt provider against beancount compiler
	@$(SHELL) ./test/mt-test-beancount.sh

test-hxsec-beancount: ## Run tests for hxsec provider against beancount compiler
	@$(SHELL) ./test/hxsec-test-beancount.sh

test-cmb-beancount: ## Run tests for cmb provider against beancount compiler
	@$(SHELL) ./test/cmb-test-beancount.sh
test-cmb-ledger: ## Run tests for cmb provider against ledger compiler
	@$(SHELL) ./test/cmb-test-ledger.sh

format: ## Format code
	@gofmt -s -w pkg
	@goimports -w pkg

install-golangci-lint:
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

clean-cache:
	@go clean -cache && go clean -modcache

lint: ## Lint GO code
	@golangci-lint run

check-format: ## Check if the format looks good.
	@go fmt ./...

before-commit-check: format check-format lint test ## Do all static checks & tests before you commit code

goreleaser-build-test: ## Goreleaser build for testing
	goreleaser build --single-target --snapshot --clean

gen-doc: ## Generate command docs by spf13/cobra
	@go run hack/generate-doc.go

clean-wasm: ## Clean wasm-dist dir
	@rm -rf ./wasm-dist

build-wasm: clean-wasm ## Build WebAssembly's version
	@mkdir -p wasm-dist
	GOOS=js GOARCH=wasm go build -o wasm-dist/double-entry-generator.wasm $(LD_FLAGS) $(CMD_DIR)/
	@cp "$$(go env GOROOT)/misc/wasm/wasm_exec.js" wasm-dist/
	@cp wasm/* wasm-dist/
	@echo "Build wasm completed! Type \`make run-wasm-server\` to run wasm."

run-wasm-server: ## Run WebAssembly in browser
	@cd wasm-dist
	# @$(BROWSER) http://127.0.0.1:2000
	@python -m http.server --directory wasm-dist --bind 127.0.0.1 2000
