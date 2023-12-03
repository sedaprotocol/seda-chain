#!/bin/bash
set -e
set -x

#
# Local Single-node Setup
#
# NOTE: Run this script from project root.
#
make build
BIN=./build/seda-chaind

$BIN tendermint unsafe-reset-all
rm -rf ~/.seda-chain
$BIN init new node0

cat $HOME/.seda-chain/config/genesis.json | jq '.app_state["gov"]["voting_params"]["voting_period"]="30s"' > $HOME/.seda-chain/config/tmp_genesis.json && mv $HOME/.seda-chain/config/tmp_genesis.json $HOME/.seda-chain/config/genesis.json
cat $HOME/.seda-chain/config/genesis.json | jq '.app_state["gov"]["params"]["voting_period"]="30s"' > $HOME/.seda-chain/config/tmp_genesis.json && mv $HOME/.seda-chain/config/tmp_genesis.json $HOME/.seda-chain/config/genesis.json
cat $HOME/.seda-chain/config/genesis.json | jq '.app_state["gov"]["params"]["expedited_voting_period"]="15s"' > $HOME/.seda-chain/config/tmp_genesis.json && mv $HOME/.seda-chain/config/tmp_genesis.json $HOME/.seda-chain/config/genesis.json
cat $HOME/.seda-chain/config/genesis.json | jq '.consensus["params"]["validator"]["pub_key_types"]=["secp256k1"]' > $HOME/.seda-chain/config/tmp_genesis.json && mv $HOME/.seda-chain/config/tmp_genesis.json $HOME/.seda-chain/config/genesis.json

$BIN keys add satoshi --keyring-backend test
ADDR=$($BIN keys show satoshi --keyring-backend test -a)
$BIN add-genesis-account $ADDR 100000000000000000seda --keyring-backend test
$BIN gentx satoshi 10000000000000000seda --keyring-backend test --chain-id sedachain

$BIN keys add acc1 --keyring-backend test
ADDR=$($BIN keys show acc1 --keyring-backend test -a)
$BIN add-genesis-account $ADDR 100000000000000000seda --keyring-backend test

$BIN collect-gentxs
$BIN start --log_level debug
