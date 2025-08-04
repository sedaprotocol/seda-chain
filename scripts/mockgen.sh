#!/usr/bin/env bash

mockgen_cmd="mockgen"

if ! [ -x "$(command -v $mockgen_cmd)" ]; then
  echo "Error: $mockgen_cmd is not installed." >&2
  exit 1
fi

mockgen_version=$($mockgen_cmd -version 2>&1 | grep -E 'v[0-9]+\.[0-9]+\.[0-9]+' || echo "unknown")
required_version="v0.5.0"

if [ "$mockgen_version" != "$required_version" ]; then
  echo "warning: required mockgen version is $required_version, but found $mockgen_version" >&2
fi

# Generate mocks for the given package
$mockgen_cmd -source=$GOPATH/pkg/mod/github.com/\!cosm\!wasm/wasmd@v0.53.0/x/wasm/types/exported_keepers.go -package testutil -destination=x/wasm-storage/keeper/testutil/wasm_keepers_mock.go
$mockgen_cmd -source=x/wasm-storage/types/expected_keepers.go -package testutil -destination=x/wasm-storage/keeper/testutil/expected_keepers_mock.go
$mockgen_cmd -source=x/pubkey/types/expected_keepers.go -package testutil -destination=x/pubkey/keeper/testutil/expected_keepers_mock.go
$mockgen_cmd -source=x/core/types/types.go -package testutil -destination=x/core/keeper/testutil/expected_keepers_mock.go
$mockgen_cmd -source=x/staking/types/expected_keepers.go -package testutil -destination=x/staking/keeper/testutil/expected_keepers_mock.go
$mockgen_cmd -source=app/abci/expected_keepers.go -package testutil -destination=app/abci/testutil/expected_keepers_mock.go
$mockgen_cmd -source=app/ante.go -package testutil -destination=app/testutil/expected_keepers_mock.go
$mockgen_cmd -source=app/utils/seda_keys.go -package testutil -destination=app/abci/testutil/seda_keys_mock.go
$mockgen_cmd -source=x/data-proxy/types/expected_keepers.go -package testutil -destination=x/data-proxy/keeper/testutil/expected_keepers_mock.go
$mockgen_cmd -source=$GOPATH/pkg/mod/github.com/aws/aws-sdk-go@v1.55.5/service/sqs/sqsiface/interface.go -package testutil -destination=plugins/indexing/pluginaws/testutil/sqs_client_mock.go
