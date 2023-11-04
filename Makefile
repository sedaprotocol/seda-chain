#!/usr/bin/env make -f
###############################################################################
##                                   Set Env Vars                            ##
###############################################################################
export CGO_ENABLED=1
export VERSION := $(shell echo $(shell git describe --tags --always --match "v*") | sed 's/^v//')
export COMMIT := $(shell git log -1 --format='%H')

###############################################################################
##                                   Set Local Vars                          ##
###############################################################################
LEDGER_ENABLED ?= true
BUILDDIR ?= $(CURDIR)/build
DOCKER := $(shell which docker)
build_tags = netgo

# If we are building ledger support
ifeq ($(LEDGER_ENABLED),true)
  ifeq ($(OS),Windows_NT)
    GCCEXE = $(shell where gcc.exe 2> NUL)
    ifeq ($(GCCEXE),)
      $(error gcc.exe not installed for ledger support, please install or set LEDGER_ENABLED=false)
    else
      build_tags += ledger
    endif
  else
    UNAME_S = $(shell uname -s)
    ifeq ($(UNAME_S),OpenBSD)
      $(warning OpenBSD detected, disabling ledger support (https://github.com/cosmos/cosmos-sdk/issues/1988))
    else
      GCC = $(shell command -v gcc 2> /dev/null)
      ifeq ($(GCC),)
        $(error gcc not installed for ledger support, please install or set LEDGER_ENABLED=false)
      else
        build_tags += ledger
      endif
    endif
  endif
endif

ifeq (secp,$(findstring secp,$(COSMOS_BUILD_OPTIONS)))
  build_tags += libsecp256k1_sdk
endif

whitespace :=
whitespace += $(whitespace)
comma := ,
build_tags_comma_sep := $(subst $(whitespace),$(comma),$(build_tags))

ldflags = -X github.com/cosmos/cosmos-sdk/version.Name=seda-chain \
		  -X github.com/cosmos/cosmos-sdk/version.AppName=seda-chaind \
		  -X github.com/cosmos/cosmos-sdk/version.Version=$(VERSION) \
		  -X github.com/cosmos/cosmos-sdk/version.Commit=$(COMMIT) \
		  -X "github.com/cosmos/cosmos-sdk/version.BuildTags=$(build_tags_comma_sep)" \
			-X github.com/tendermint/tendermint/version.TMCoreSemVer=$(TMVERSION)

# DB Selection
ifeq ($(ENABLE_ROCKSDB),true)
  BUILD_TAGS += rocksdb
  test_tags += rocksdb
endif

# DB backend selection
ifeq (rocksdb,$(findstring rocksdb,$(COSMOS_BUILD_OPTIONS)))
  ifneq ($(ENABLE_ROCKSDB),true)
    $(error Cannot use RocksDB backend unless ENABLE_ROCKSDB=true)
  endif
endif

# For building statically linked binaries
ifeq ($(LINK_STATICALLY),true)
	ldflags += -linkmode=external -extldflags "-Wl,-z,muldefs -static"
endif

ifeq (,$(findstring nostrip,$(COSMOS_BUILD_OPTIONS)))
  ldflags += -w -s
endif
ldflags += $(LDFLAGS)
ldflags := $(strip $(ldflags))

build_tags += $(BUILD_TAGS)
build_tags := $(strip $(build_tags))

BUILD_FLAGS := -tags "$(build_tags)" -ldflags '$(ldflags)'
# check for nostrip option
ifeq (,$(findstring nostrip,$(COSMOS_BUILD_OPTIONS)))
  BUILD_FLAGS += -trimpath
endif

# Check for debug option
ifeq (debug,$(findstring debug,$(COSMOS_BUILD_OPTIONS)))
  BUILD_FLAGS += -gcflags "all=-N -l"
endif

# default make command
all: clean go.sum fmt lint build

###############################################################################
##                                   Build                                   ##
###############################################################################

BUILD_TARGETS := build install

build: BUILD_ARGS=-o $(BUILDDIR)/

$(BUILD_TARGETS): go.sum $(BUILDDIR)/
	@go $@ -mod=readonly $(BUILD_FLAGS) $(BUILD_ARGS) ./...

$(BUILDDIR)/:
	@mkdir -p $(BUILDDIR)/

clean:
	@echo "--> Cleaning..."
	@rm -rf $(BUILDDIR)/** 

.PHONY: build clean

###############################################################################
###                          Tools & Dependencies                           ###
###############################################################################

go.sum: go.mod
	@echo "Ensure dependencies have not been modified ..." >&2
	@go mod verify
	@go mod tidy

fmt:
	@echo "--> Running go fmt"
	$(shell go fmt ./...)

###############################################################################
###                                Linting                                  ###
###############################################################################

golangci_version=v1.53.3

lint-install:
	@echo "--> Installing golangci-lint $(golangci_version)"
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(golangci_version)

lint:
	@echo "--> Running linter"
	@./scripts/go-lint-all.bash --timeout=15m

.PHONY: lint lint-install

###############################################################################
###                                Protobuf                                 ###
###############################################################################

proto-all: proto-gen proto-format proto-lint

proto-dep-install:
	@go install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway@latest
	@go install github.com/cosmos/gogoproto/protoc-gen-gocosmos@latest
	@go install github.com/cosmos/gogoproto/protoc-gen-gogo@latest

proto-gen:
	@echo "Generating Protobuf files"
	@./scripts/proto_gen.sh

proto-fmt:
	@echo "Formatting Protobuf files"
	@find ./ -name "*.proto" -exec clang-format -i {} \;

proto-lint:
	@echo "Linting Protobuf files"
	@buf lint --error-format=json ./proto

proto-update-deps:
	@echo "Updating Protobuf dependencies"
	@buf mod update ./proto

.PHONY: proto-gen proto-lint proto-update-deps

###############################################################################
##                                   Tests                                   ##
###############################################################################

PACKAGES_UNIT=$(shell go list ./... | grep -v /e2e)
PACKAGES_E2E=$(shell go list ./... | grep /e2e)

TEST_PACKAGES=./...
TEST_TARGETS := test-unit test-unit-cover test-unit-race
TEST_COVERAGE_PROFILE=coverage.txt

UNIT_TEST_TAGS = norace
TEST_RACE_TAGS = ""

# runs all tests
test: test-unit
test-race: test-unit-race

test-unit: ARGS=-timeout=10m -tags='$(UNIT_TEST_TAGS)'
test-unit: TEST_PACKAGES=$(PACKAGES_UNIT)
test-unit-cover: ARGS=-timeout=10m -tags='$(UNIT_TEST_TAGS)' -coverprofile=$(TEST_COVERAGE_PROFILE) -covermode=atomic
test-unit-cover: TEST_PACKAGES=$(PACKAGES_UNIT)
test-unit-race: ARGS=-timeout=10m -race -tags='$(TEST_RACE_TAGS)'
test-unit-race: TEST_PACKAGES=$(PACKAGES_UNIT)
test-e2e: ARGS=-timeout=10m -v
test-e2e: TEST_PACKAGES=$(PACKAGES_E2E)
$(TEST_TARGETS): run-tests

run-tests:
ifneq (,$(shell which tparse 2>/dev/null))
	@echo "--> Running tests"
	@go test -mod=readonly -json $(ARGS) $(TEST_PACKAGES) | tparse
else
	@echo "--> Running tests"
	@go test -mod=readonly $(ARGS) $(TEST_PACKAGES)
endif

cover-html: test-unit-cover
	@echo "--> Opening in the browser"
	@go tool cover -html=$(TEST_COVERAGE_PROFILE)

.PHONY: cover-html run-tests $(TEST_TARGETS) test test-race


###############################################################################
###                                Docker                                   ###
###############################################################################

docker-build-e2e:
	@docker build -t sedaprotocol/seda-chaind-e2e -f dockerfiles/Dockerfile.e2e .

.PHONY: docker-build-e2e


###############################################################################
###                                Release                                  ###
###############################################################################

GO_VERSION=1.21
GORELEASER_IMAGE := ghcr.io/goreleaser/goreleaser-cross:v$(GO_VERSION)
COSMWASM_VERSION := $(shell go list -m github.com/CosmWasm/wasmvm | sed 's/.* //')
ifdef GITHUB_TOKEN
release:
	docker run \
		--rm \
		-e GITHUB_TOKEN=$(GITHUB_TOKEN) \
		-e COSMWASM_VERSION=$(COSMWASM_VERSION) \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v `pwd`:/go/src/seda-chaind \
		-w /go/src/seda-chaind \
		$(GORELEASER_IMAGE) \
		release \
		--clean
else
release:
	@echo "Error: GITHUB_TOKEN variable required to 'make release'."

endif

release-dry-run:
	@docker run \
		--rm \
		-e COSMWASM_VERSION=$(COSMWASM_VERSION) \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v `pwd`:/go/src/seda-chaind \
		-w /go/src/seda-chaind \
		$(GORELEASER_IMAGE) \
		release \
		--clean \
		--skip=publish

release-snapshot:
	@docker run \
		--rm \
		-e COSMWASM_VERSION=$(COSMWASM_VERSION) \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v `pwd`:/go/src/seda-chaind \
		-w /go/src/seda-chaind \
		$(GORELEASER_IMAGE) \
		release \
		--clean \
		--snapshot \
		--skip-validate\
		--skip-publish

.PHONY: release release-dry-run release-snapshot

###############################################################################
###                                Docker                                  ###
###############################################################################
RUNNER_BASE_IMAGE_DISTROLESS := gcr.io/distroless/static-debian11
RUNNER_BASE_IMAGE_ALPINE := alpine:3.17

docker-static-build:
	@DOCKER_BUILDKIT=1 docker build \
		-t seda-chain/seda-chaind-static-distroless \
		--build-arg GO_VERSION=$(GO_VERSION) \
		--build-arg RUNNER_IMAGE=$(RUNNER_BASE_IMAGE_DISTROLESS) \
		--build-arg GIT_VERSION=$(VERSION) \
		--build-arg GIT_COMMIT=$(COMMIT) \
		-f $(CURDIR)/dockerfiles/Dockerfile.static .

docker-static-build-alpine:
	@DOCKER_BUILDKIT=1 docker build \
		-t seda-chain/seda-chaind-static-alpine \
		--build-arg GO_VERSION=$(GO_VERSION) \
		--build-arg RUNNER_IMAGE=$(RUNNER_BASE_IMAGE_ALPINE) \
		--build-arg GIT_VERSION=$(VERSION) \
		--build-arg GIT_COMMIT=$(COMMIT) \
		-f $(CURDIR)/dockerfiles/Dockerfile.static .

.PHONY: docker-static-build docker-static-build-alpine