#!/usr/bin/env bash

set -eo pipefail

mkdir -p ./tmp-swagger-gen
cd proto
seda_proto_dirs=$(find ./sedachain -path -prune -o -name '*.proto' -print0 | xargs -0 -n1 dirname | sort | uniq)

SDK_VERSION=$(go list -m github.com/cosmos/cosmos-sdk | sed 's/.* //')
COSMOS_PROTO_DIR=$(go env GOPATH)/pkg/mod/github.com/cosmos/cosmos-sdk@$SDK_VERSION/proto
cosmos_proto_dirs=$(find $COSMOS_PROTO_DIR/cosmos -path -prune -o -name '*.proto' -print0 | xargs -0 -n1 dirname | sort | uniq)

proto_dirs=("${seda_proto_dirs[@]}" "${cosmos_proto_dirs[@]}")
echo $proto_dirs
for dir in $proto_dirs; do
  # generate swagger files (filter query files)
  query_file=$(find "${dir}" -maxdepth 1 \( -name 'query.proto' -o -name 'service.proto' \))
  if [[ ! -z "$query_file" ]]; then
    buf generate --template buf.gen.swagger.yaml $query_file
  fi
done

cd ..
# combine swagger files
# uses nodejs package `swagger-combine`.
# all the individual swagger files need to be configured in `config.json` for merging
swagger-combine ./client/docs/config.json -o ./client/docs/swagger-ui/swagger.yaml -f yaml --continueOnConflictingPaths true --includeDefinitions true

# clean swagger files
# rm -rf ./tmp-swagger-gen
