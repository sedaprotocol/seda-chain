#!/usr/bin/env bash
# Tell script to exit immediately if any cmd fails
set -e

BIN="seda-chaind"
KEY_NAME="${KEY_NAME:-default_key}"
STAKE_AMOUNT="$1aseda"

source common.sh

validator_pub_key=$($BIN tendermint show-validator --home=$HOME/.seda-chain)

chain_id=$(cat $HOME/.seda-chain/config/genesis.json | jq .chain_id | tr -d '"')
auth_seda_chaind_command tx unbond $validator_pub_key --amount=$STAKE_AMOUNT --pubkey=$validator_pub_key --moniker=$MONIKER --commission-rate=0.10 --commission-max-rate=0.20 --commission-max-change-rate=0.01 --gas=auto --gas-adjustment=1.2 --gas-prices=0.0025aseda --from=$KEY_NAME --min-self-delegation=1 --yes --home=$HOME/.seda-chain --chain-id=$chain_id

