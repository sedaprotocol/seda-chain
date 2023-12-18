#!/bin/bash
set -e
set -x

BIN=./build/seda-chaind
CONFIG_PATH=$HOME/.seda-chain/config

function add_key_and_account() {
    local name=$1
    local amount=$2
    $BIN keys add $name --keyring-backend test
    $BIN add-genesis-account $name $amount --keyring-backend test
}

#
# Local Single-node Setup
#
# NOTE: Run this script from project root.
#

# build the binary
make build

# reset the chain
$BIN tendermint unsafe-reset-all
rm -rf ~/.seda-chain || true

# configure seda-chaind
$BIN config set client chain-id sedachain

# initialize the chain
$BIN init node0 --default-denom aseda

cat $HOME/.seda-chain/config/genesis.json | jq '.app_state["gov"]["voting_params"]["voting_period"]="30s"' > $HOME/.seda-chain/config/tmp_genesis.json && mv $HOME/.seda-chain/config/tmp_genesis.json $HOME/.seda-chain/config/genesis.json
cat $HOME/.seda-chain/config/genesis.json | jq '.app_state["gov"]["params"]["voting_period"]="30s"' > $HOME/.seda-chain/config/tmp_genesis.json && mv $HOME/.seda-chain/config/tmp_genesis.json $HOME/.seda-chain/config/genesis.json
cat $HOME/.seda-chain/config/genesis.json | jq '.app_state["gov"]["params"]["expedited_voting_period"]="15s"' > $HOME/.seda-chain/config/tmp_genesis.json && mv $HOME/.seda-chain/config/tmp_genesis.json $HOME/.seda-chain/config/genesis.json
cat $HOME/.seda-chain/config/genesis.json | jq '.consensus_params["block"]["max_gas"]="100000000"' > $HOME/.seda-chain/config/tmp_genesis.json && mv $HOME/.seda-chain/config/tmp_genesis.json $HOME/.seda-chain/config/genesis.json

# update genesis
add_key_and_account "satoshi" "100000000000000000seda"
add_key_and_account "acc1" "100000000000000000seda"

# create a default validator
$BIN gentx satoshi 10000000000000000seda --keyring-backend test

# collect genesis txns
$BIN collect-gentxs

# start the chain
$BIN start --log_level debug || echo "Failed to start the chain"