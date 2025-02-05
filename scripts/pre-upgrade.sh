#!/bin/bash
set -ex

# Configure locations of old and new binaries.
OLD_BIN=$(git rev-parse --show-toplevel)/scripts/upgrade-prep-v1.0.0/sedad_old
NEW_BIN=$(git rev-parse --show-toplevel)/scripts/upgrade-prep-v1.0.0/sedad_new
GENESIS_FILE=$(git rev-parse --show-toplevel)/scripts/upgrade-prep-v1.0.0/genesis.json
PRIV_VAL_KEY=$(git rev-parse --show-toplevel)/scripts/upgrade-prep-v1.0.0/priv_validator_key.json

# Other configurations that rarely change
TMP_COSMOVISOR_DIR=tmp-cosmovisor
CHAIN_ID=seda-1

export DAEMON_HOME=$HOME/.sedad
export DAEMON_NAME=sedad
export DAEMON_ALLOW_DOWNLOAD_BINARIES=false 
export DAEMON_RESTART_AFTER_UPGRADE=true
export SEDA_ALLOW_UNENCRYPTED_KEYS=true

rm -rf ~/.sedad

$OLD_BIN init node0 --chain-id $CHAIN_ID --home $DAEMON_HOME

cp $GENESIS_FILE $DAEMON_HOME/config/genesis.json
cp $PRIV_VAL_KEY $DAEMON_HOME/config/priv_validator_key.json

mkdir -p $TMP_COSMOVISOR_DIR/genesis/bin
cp $OLD_BIN $TMP_COSMOVISOR_DIR/genesis/bin/sedad
mkdir -p $TMP_COSMOVISOR_DIR/upgrades/v1.0.0/bin
cp $NEW_BIN $TMP_COSMOVISOR_DIR/upgrades/v1.0.0/bin/sedad

cp -R $TMP_COSMOVISOR_DIR $DAEMON_HOME/cosmovisor
cosmovisor run start --home $DAEMON_HOME
