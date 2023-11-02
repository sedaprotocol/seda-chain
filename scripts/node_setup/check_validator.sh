#!/usr/bin/env bash
# Tell script to exit immediately if any cmd fails
set -e

BIN="seda-chaind"
KEY_NAME="${KEY_NAME:-default_key}"

source common.sh

validator_address=$(auth_seda_chaind_command keys show $KEY_NAME --bech val | grep "address:" | awk '{print $3}')
output=$(seda-chaind query staking validator $validator_address 2>&1)
echo "$output" | grep "error:" || echo "$output"

