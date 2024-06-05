#!/usr/bin/env bash

mockgen_cmd="mockgen"

if ! [ -x "$(command -v $mockgen_cmd)" ]; then
  echo "Error: $mockgen_cmd is not installed." >&2
  exit 1
fi

# Generate mocks for the given package
$mockgen_cmd -source=$GOPATH/pkg/mod/github.com/\!cosm\!wasm/wasmd@v0.50.0/x/wasm/types/exported_keepers.go -package testutil -destination x/wasm-storage/keeper/testutil/wasm_keepers_mock.go
$mockgen_cmd -source=x/wasm-storage/types/expected_keepers.go -package testutil -destination x/wasm-storage/keeper/testutil/expected_keepers_mock.go
