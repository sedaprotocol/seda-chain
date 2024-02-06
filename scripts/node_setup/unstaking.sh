#!/usr/bin/env bash
# Tell script to exit immediately if any cmd fails
set -e

BIN="sedad"
KEY_NAME="${KEY_NAME:-default_key}"
STAKE_AMOUNT="$1aseda"

source common.sh

validator_address=$(auth_seda_chaind_command keys show $KEY_NAME --bech val | grep "address:" | awk '{print $3}')  

chain_id=$(cat $HOME/.seda/config/genesis.json | jq .chain_id | tr -d '"')
auth_seda_chaind_command tx staking unbond $validator_address  --gas=auto --gas-adjustment=1.2 --gas-prices=0.0025aseda --from=$KEY_NAME --yes --home=$HOME/.seda --chain-id=$chain_id $STAKE_AMOUNT