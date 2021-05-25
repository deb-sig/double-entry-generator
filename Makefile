# Copyright 2019 The Caicloud Authors.
#
# The old school Makefile, following are required targets. The Makefile is written
# to allow building multiple binaries. You are free to add more targets or change
# existing implementations, as long as the semantics are preserved.
# Ref https://github.com/caicloud/golang-template-project/blob/master/Makefile

# This repo's root import path (under GOPATH).
ROOT := github.com/gaocegege/double-entry-generator

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

# All targets.
.PHONY: lint test build container push

build: build-local

build-local:
	@for target in $(TARGETS); do                                                      \
	  go build -i -v -o $(OUTPUT_DIR)/$${target}                                       \
	  -ldflags "-s -w -X $(ROOT)/pkg/version.VERSION=$(VERSION)                        \
	    -X $(ROOT)/pkg/version.REPOROOT=$(ROOT)                                        \
		-X $(ROOT)/pkg/version.COMMIT=$(GIT_COMMIT)"                                   \
	  $(CMD_DIR)/;                                                                     \
	done

install: build
	@install ./bin/double-entry-generator /usr/local/bin
	@echo "Installed double-entry-generator at /usr/local/bin/double-entry-generator !"
