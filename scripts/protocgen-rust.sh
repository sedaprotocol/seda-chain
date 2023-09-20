#!/usr/bin/env bash

set -eo pipefail

echo "Generating Rust proto code"
cd proto
proto_dirs=$(find ./sedachain -path -prune -o -name '*.proto' -print0 | xargs -0 -n1 dirname | sort | uniq)
for dir in $proto_dirs; do
  buf generate --template buf.gen.rust.yaml --path $dir
done
