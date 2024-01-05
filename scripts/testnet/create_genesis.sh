#!/bin/bash
set -euxo pipefail

#
# This script accomplishes the following:
#   - Add genesis accounts
#   - Create node key and validator file for each given node
#   - Create and collect gentxs
#
# The resulting files are placed in $NODE_DIR
#
source config.sh

#
#   PRELIMINARY PROCESS
#
rm -rf $HOME_DIR
rm -rf $NODE_DIR

#
#   CREATE GENESIS AND ADJUST GOV PARAMETERS
#
$BIN init node0 --chain-id $CHAIN_ID --default-denom aseda

cat $HOME/.seda-chain/config/genesis.json | jq '.app_state["gov"]["voting_params"]["voting_period"]="180s"' > $HOME/.seda-chain/config/tmp_genesis.json && mv $HOME/.seda-chain/config/tmp_genesis.json $HOME/.seda-chain/config/genesis.json
cat $HOME/.seda-chain/config/genesis.json | jq '.app_state["gov"]["params"]["voting_period"]="180s"' > $HOME/.seda-chain/config/tmp_genesis.json && mv $HOME/.seda-chain/config/tmp_genesis.json $HOME/.seda-chain/config/genesis.json
cat $HOME/.seda-chain/config/genesis.json | jq '.app_state["gov"]["params"]["max_deposit_period"]="180s"' > $HOME/.seda-chain/config/tmp_genesis.json && mv $HOME/.seda-chain/config/tmp_genesis.json $HOME/.seda-chain/config/genesis.json
cat $HOME/.seda-chain/config/genesis.json | jq '.consensus["params"]["validator"]["pub_key_types"]=["secp256k1"]' > $HOME/.seda-chain/config/tmp_genesis.json && mv $HOME/.seda-chain/config/tmp_genesis.json $HOME/.seda-chain/config/genesis.json
cat $HOME/.seda-chain/config/genesis.json | jq '.consensus_params["block"]["max_gas"]="100000000"' > $HOME/.seda-chain/config/tmp_genesis.json && mv $HOME/.seda-chain/config/tmp_genesis.json $HOME/.seda-chain/config/genesis.json

# TO-DO?
# - chain id
# - launch time

#
#   ADD GENESIS ACCOUNTS
#
for i in ${!GENESIS_ADDRESSES[@]}; do
    $BIN add-genesis-account ${GENESIS_ADDRESSES[$i]} 100000000000000000seda
done

set +u
if [ ! -z "$SATOSHI" ]; then
    $BIN add-genesis-account $SATOSHI 10000000000000000000seda
fi
if [ ! -z "$FAUCET" ]; then
    $BIN add-genesis-account $FAUCET 1000000000000000000seda
fi
set -u

#
# CREATE NODE KEY, VALIDATOR KEY, AND GENTX FOR EACH NODE
#
GENTX_DIR=$NODE_DIR/gentx
mkdir -p $GENTX_DIR

for i in ${!MONIKERS[@]}; do
    INDIVIDUAL_VAL_HOME_DIR=$NODE_DIR/${MONIKERS[$i]}
    INDIVIDUAL_VAL_CONFIG_DIR="$INDIVIDUAL_VAL_HOME_DIR/config"

    $BIN init ${MONIKERS[$i]} --home $INDIVIDUAL_VAL_HOME_DIR  --chain-id $CHAIN_ID --default-denom aseda
    $BIN keys add ${MONIKERS[$i]} --keyring-backend=test --home $INDIVIDUAL_VAL_HOME_DIR

    VALIDATOR_ADDRESS=$($BIN keys show ${MONIKERS[$i]} --keyring-backend test --home $INDIVIDUAL_VAL_HOME_DIR -a)

    # to create their gentx
    $BIN add-genesis-account $VALIDATOR_ADDRESS 500000000000000000seda --home $INDIVIDUAL_VAL_HOME_DIR
    # to output geneis file
    $BIN add-genesis-account $VALIDATOR_ADDRESS 500000000000000000seda

    $BIN gentx ${MONIKERS[$i]} ${SELF_DELEGATION_AMOUNTS[$i]} --moniker=${MONIKERS[$i]} --keyring-backend=test --home $INDIVIDUAL_VAL_HOME_DIR --ip=${IPS[$i]} --chain-id $CHAIN_ID

    cp -a $INDIVIDUAL_VAL_CONFIG_DIR/gentx/. $GENTX_DIR
done

cp -r $GENTX_DIR $HOME_CONFIG_DIR
$BIN collect-gentxs --home $HOME_DIR
cp $HOME_CONFIG_DIR/genesis.json $NODE_DIR
