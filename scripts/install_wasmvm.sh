#!/usr/bin/env bash

ARCH=$(uname -m)
WASMVM_VERSION=$(go list -m github.com/CosmWasm/wasmvm | sed 's/.* //')

wget https://github.com/CosmWasm/wasmvm/releases/download/$WASMVM_VERSION/libwasmvm.$ARCH.so -O /lib/libwasmvm.$ARCH.so && \
  # verify checksum
  wget https://github.com/CosmWasm/wasmvm/releases/download/$WASMVM_VERSION/checksums.txt -O /tmp/checksums.txt && \
  sha256sum /lib/libwasmvm.$ARCH.so | grep $(cat /tmp/checksums.txt | grep libwasmvm.$ARCH.so | cut -d ' ' -f 1)