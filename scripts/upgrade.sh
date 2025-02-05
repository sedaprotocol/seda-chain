#!/bin/bash 
set -ex

export DAEMON_HOME=$HOME/.sedad

OLD_BIN=$HOME/.sedad/cosmovisor/genesis/bin/sedad
UPGRADE_NAME=v1.0.0
CHAIN_ID=seda-1
PROPOSAL_JSON=$(git rev-parse --show-toplevel)/scripts/upgrade-prep-v1.0.0/proposal.json
SEED="cause oblige possible opera pluck leaf mirror start pig glare decorate gauge blast lava empower eagle keen renew remain surge culture bonus fabric twist"

echo $SEED | $OLD_BIN keys add satoshi --keyring-backend test --recover

$OLD_BIN tx gov submit-proposal $PROPOSAL_JSON \
    --from satoshi --keyring-backend test --home $DAEMON_HOME \
    --gas auto --gas-prices 10000000000aseda --gas-adjustment 1.8 \
    --chain-id $CHAIN_ID --yes | $OLD_BIN q wait-tx 

$OLD_BIN tx gov vote 1 yes \
    --from satoshi --keyring-backend test --home $DAEMON_HOME \
    --gas auto --gas-prices 10000000000aseda --gas-adjustment 1.8 \
    --chain-id $CHAIN_ID --yes

# To query the upgrade plan, run:
# $OLD_BIN query upgrade plan
