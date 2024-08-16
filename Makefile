#!/usr/bin/env make -f
##########################################################################
##                              Set Env Vars                            ##
##########################################################################
export CGO_ENABLED=1
export VERSION := $(shell echo $(shell git describe --tags --always --match "v*"))
export COMMIT := $(shell git log -1 --format='%H')

#########################################################################
##                             Set Local Vars                          ##
#########################################################################
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
		  -X github.com/cosmos/cosmos-sdk/version.AppName=sedad \
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

# Rosetta support
ifeq ($(ENABLE_ROSETTA),true)
  BUILD_TAGS += rosetta
endif

# For building statically linked binaries
ifeq ($(LINK_STATICALLY),true)
  ifeq ($(ENABLE_ROSETTA),true)
    $(error Cannot link statically when Rosetta is enabled)
  endif
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

build-plugin:
	@go build -o $(BUILDDIR)/plugin ./plugins/indexing/plugin.go

build-plugin-dev:
	@go build --tags dev -o $(BUILDDIR)/plugin ./plugins/indexing/plugin.go

clean:
	@echo "--> Cleaning..."
	@rm -rf $(BUILDDIR)/** 

.PHONY: build build-plugin build-plugin-dev clean

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

proto-all: proto-gen proto-fmt proto-lint

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


############################################################################
###                             Documentation                            ###
############################################################################

protoVer=0.14.0
protoImageName=ghcr.io/cosmos/proto-builder:$(protoVer)
protoImage=$(DOCKER) run --rm -v $(CURDIR):/workspace --workdir /workspace $(protoImageName)
cosmosVer=$(shell go list -m github.com/cosmos/cosmos-sdk | sed 's/.* //')

SWAGGER_DIR=./swagger-proto
THIRD_PARTY_DIR=$(SWAGGER_DIR)/third_party
BINDIR ?= $(GOPATH)/bin
STATIK = $(BINDIR)/statik

proto-swagger-gen:
	@make clean
	@echo "Downloading Protobuf dependencies"
	@make proto-download-deps
	@echo "Generating Protobuf Swagger"
	@$(protoImage) sh ./scripts/protoc-swagger-gen.sh
	@echo "Generating static files for swagger docs"
	$(MAKE) update-swagger-docs

proto-download-deps:
	mkdir -p "$(THIRD_PARTY_DIR)/cosmos_tmp" && \
	cd "$(THIRD_PARTY_DIR)/cosmos_tmp" && \
	git init && \
	git remote add origin "https://github.com/cosmos/cosmos-sdk.git" && \
	git config core.sparseCheckout true && \
	printf "proto\nthird_party\n" > .git/info/sparse-checkout && \
	git fetch --depth=1 origin "$(cosmosVer)" && \
	git checkout FETCH_HEAD && \
	rm -f ./proto/buf.* && \
	mv ./proto/* ..
	rm -rf "$(THIRD_PARTY_DIR)/cosmos_tmp"

	mkdir -p "$(THIRD_PARTY_DIR)/ibc_tmp" && \
	cd "$(THIRD_PARTY_DIR)/ibc_tmp" && \
	git init && \
	git remote add origin "https://github.com/cosmos/ibc-go.git" && \
	git config core.sparseCheckout true && \
	printf "proto\n" > .git/info/sparse-checkout && \
	git pull origin main && \
	rm -f ./proto/buf.* && \
	mv ./proto/* ..
	rm -rf "$(THIRD_PARTY_DIR)/ibc_tmp"

	mkdir -p "$(THIRD_PARTY_DIR)/cosmos_proto_tmp" && \
	cd "$(THIRD_PARTY_DIR)/cosmos_proto_tmp" && \
	git init && \
	git remote add origin "https://github.com/cosmos/cosmos-proto.git" && \
	git config core.sparseCheckout true && \
	printf "proto\n" > .git/info/sparse-checkout && \
	git pull origin main && \
	rm -f ./proto/buf.* && \
	mv ./proto/* ..
	rm -rf "$(THIRD_PARTY_DIR)/cosmos_proto_tmp"

	mkdir -p "$(THIRD_PARTY_DIR)/gogoproto" && \
	curl -SSL https://raw.githubusercontent.com/cosmos/gogoproto/main/gogoproto/gogo.proto > "$(THIRD_PARTY_DIR)/gogoproto/gogo.proto"

	mkdir -p "$(THIRD_PARTY_DIR)/google/api" && \
	curl -sSL https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/annotations.proto > "$(THIRD_PARTY_DIR)/google/api/annotations.proto"
	curl -sSL https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/http.proto > "$(THIRD_PARTY_DIR)/google/api/http.proto"

	mkdir -p "$(THIRD_PARTY_DIR)/cosmos/ics23/v1" && \
	curl -sSL https://raw.githubusercontent.com/cosmos/ics23/master/proto/cosmos/ics23/v1/proofs.proto > "$(THIRD_PARTY_DIR)/cosmos/ics23/v1/proofs.proto"

update-swagger-docs: statik
	$(BINDIR)/statik -src=client/docs/swagger-ui -dest=client/docs -f -m
	@if [ -n "$(git status --porcelain)" ]; then \
		echo "\033[91mSwagger docs are out of sync!!!\033[0m";\
		exit 1;\
	else \
		echo "\033[92mSwagger docs are in sync\033[0m";\
	fi

statik: $(STATIK)
$(STATIK):
	@echo "Installing statik..."
	go install github.com/rakyll/statik@v0.1.6

.PHONY: proto-gen proto-fmt proto-lint proto-update-deps proto-swagger-gen update-swagger-docs docs

###############################################################################
##                                   Tests                                   ##
###############################################################################

PACKAGES_UNIT=$(shell go list ./... | grep -v /e2e)
PACKAGES_E2E=$(shell go list ./... | grep /e2e)

TEST_PACKAGES=./...
TEST_TARGETS := test-unit test-unit-cover test-unit-race test-e2e
TEST_COVERAGE_PROFILE=coverage.txt

UNIT_TEST_TAGS = norace
TEST_RACE_TAGS = ""

test-unit: ARGS=-timeout=10m -tags='$(UNIT_TEST_TAGS)'
test-unit: TEST_PACKAGES=$(PACKAGES_UNIT)
test-unit-cover: ARGS=-timeout=10m -tags='$(UNIT_TEST_TAGS)' -coverprofile=$(TEST_COVERAGE_PROFILE) -covermode=atomic
test-unit-cover: TEST_PACKAGES=$(PACKAGES_UNIT)
test-unit-race: ARGS=-timeout=10m -race -tags='$(TEST_RACE_TAGS)'
test-unit-race: TEST_PACKAGES=$(PACKAGES_UNIT)
test-e2e: docker-build-e2e
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

ifdef GITHUB_TOKEN
docker-build-e2e:
	@docker build \
		--build-arg GITHUB_TOKEN=$(GITHUB_TOKEN) \
		-t sedaprotocol/sedad-e2e \
		-f dockerfiles/Dockerfile.e2e .
else
docker-build-e2e:
	@echo "Error: GITHUB_TOKEN variable required to build e2e image"
endif

.PHONY: cover-html run-tests $(TEST_TARGETS) test test-race docker-build-e2e

###############################################################################
###                             Simulation Tests                            ###
###############################################################################

CURRENT_DIR = $(shell pwd)

test-sim-determinism:
	@echo "Running determinism test..."
	@cd ${CURRENT_DIR}/simulation && go test -mod=readonly -run TestAppStateDeterminism -Enabled=true \
		-NumBlocks=100 -BlockSize=200 -Commit=true -Period=0 -v -timeout 24h

test-sim-export-import:
	@echo "Running export-import test..."
	@cd ${CURRENT_DIR}/simulation && go test -mod=readonly -run TestAppExportImport -Enabled=true \
		-NumBlocks=100 -BlockSize=200 -Commit=true -Period=0 -v -timeout 24h

test-sim-after-import:
	@echo "Running simulation-after-import test..."
	@cd ${CURRENT_DIR}/simulation && go test -mod=readonly -run TestAppSimulationAfterImport -Enabled=true \
		-NumBlocks=100 -BlockSize=200 -Commit=true -Period=0 -v -timeout 24h

.PHONY: test-sim-determinism test-sim-export-import test-sim-after-import

SIM_NUM_BLOCKS ?= 500
SIM_BLOCK_SIZE ?= 200
SIM_COMMIT ?= true

test-sim-benchmark:
	@echo "Running application benchmark for numBlocks=$(SIM_NUM_BLOCKS), blockSize=$(SIM_BLOCK_SIZE). This may take awhile!"
	@cd ${CURRENT_DIR}/simulation && go test -mod=readonly -run=^$$ $(.) -bench ^BenchmarkSimulation$$  \
		-Enabled=true -NumBlocks=$(SIM_NUM_BLOCKS) -BlockSize=$(SIM_BLOCK_SIZE) -Commit=$(SIM_COMMIT) -timeout 24h

.PHONY: test-sim-benchmark

###############################################################################
###                                interchaintest                           ###
###############################################################################

ictest-sdk-commands: rm-testcache
	cd interchaintest && go test -race -v -run TestCoreSDKCommands .

ictest-sdk-boundaries: rm-testcache
	cd interchaintest && go test -race -v -run TestSDKBoundaries .

ictest-chain-start: rm-testcache
	cd interchaintest && go test -race -v -run TestChainStart .

ictest-state-sync: rm-testcache
	cd interchaintest && go test -race -v -run TestStateSync .

ictest-ibc-xfer: rm-testcache
	cd interchaintest && go test -race -v -run TestIBCTransfer .

ictest-packet-forward-middleware: rm-testcache
	cd interchaintest && go test -race -v -run TestPacketForwardMiddleware .

ictest-ibc-ica: rm-testcache
	cd interchaintest && go test -race -v -run TestInterchainAccounts .

rm-testcache:
	go clean -testcache

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
		-v `pwd`:/go/src/sedad \
		-w /go/src/sedad \
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
		-v `pwd`:/go/src/sedad \
		-w /go/src/sedad \
		$(GORELEASER_IMAGE) \
		release \
		--clean \
		--skip=publish

release-snapshot:
	@docker run \
		--rm \
		-e COSMWASM_VERSION=$(COSMWASM_VERSION) \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v `pwd`:/go/src/sedad \
		-w /go/src/sedad \
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
		-t seda-chain/sedad-static-distroless \
		--build-arg GO_VERSION=$(GO_VERSION) \
		--build-arg RUNNER_IMAGE=$(RUNNER_BASE_IMAGE_DISTROLESS) \
		--build-arg GIT_VERSION=$(VERSION) \
		--build-arg GIT_COMMIT=$(COMMIT) \
		-f $(CURDIR)/dockerfiles/Dockerfile.static .

docker-static-build-alpine:
	@DOCKER_BUILDKIT=1 docker build \
		-t seda-chain/sedad-static-alpine \
		--build-arg GO_VERSION=$(GO_VERSION) \
		--build-arg RUNNER_IMAGE=$(RUNNER_BASE_IMAGE_ALPINE) \
		--build-arg GIT_VERSION=$(VERSION) \
		--build-arg GIT_COMMIT=$(COMMIT) \
		-f $(CURDIR)/dockerfiles/Dockerfile.static .

.PHONY: docker-static-build docker-static-build-alpine
