#!/bin/bash
set -euxo pipefail

#
# This script accomplishes the following:
#   - Add genesis accounts
#   - Create node key and validator file for each given node
#   - Create and collect gentxs
#
# The resulting files are placed in $OUT_DIR
#

#
#   PARAMETERS
#
OUT_DIR=./nodes # output directory
BIN=../../build/seda-chaind # chain binary executable on your machine
HOME_DIR=$HOME/.seda-chain # chain directory
CONFIG_DIR=$HOME_DIR/config # chain config directory
# validators
IPS=(
    "18.169.59.167"
    "35.178.98.62"
)
MONIKERS=(
    "node0"
    "node1"
)
SELF_DELEGATION_AMOUNTS=(
    "30000000000000000seda"
    "10000000000000000seda"
)
# genesis acoounts addresses
ADDR1=seda19gqrkdjhju0txurteag8vle90p09a5r5dd78rp
ADDR2=seda1gnes565n2vhldm2eerm5fcuwz2mpcadvqnvped
ADDR3=seda1wr0la8asy5wg9ja83rvdy36cmp4qrztypytdl7
ADDR4=seda154aany5fudkp9mncekupm3hwr7w3da3dv79c4k
ADDR5=seda15yfxudv7ek8m6ecxt4u9v5a677yhm3d662z3fg
ADDR6=seda1c3czshqflpxs9eyns9r906gk9s9xfcpsf7rcac
ADDR7=seda1uvraznfum5zc2tke5vu3hcj9n7a4ndcv533gnr
ADDR8=seda1z3ecw3k2asd5gd82v7m78y6u5y5vm7xnp46lf2
# CHAIN_ID=seda-testnet
# GENESIS_TIME=


#
#   PRELIMINARY PROCESS
#
rm -rf $HOME_DIR
rm -rf $OUT_DIR

#
#   CREATE GENESIS AND ADJUST GOV PARAMETERS
#
$BIN init new node0

cat $HOME/.seda-chain/config/genesis.json | jq '.app_state["gov"]["voting_params"]["voting_period"]="180s"' > $HOME/.seda-chain/config/tmp_genesis.json && mv $HOME/.seda-chain/config/tmp_genesis.json $HOME/.seda-chain/config/genesis.json
cat $HOME/.seda-chain/config/genesis.json | jq '.app_state["gov"]["params"]["voting_period"]="180s"' > $HOME/.seda-chain/config/tmp_genesis.json && mv $HOME/.seda-chain/config/tmp_genesis.json $HOME/.seda-chain/config/genesis.json
cat $HOME/.seda-chain/config/genesis.json | jq '.app_state["gov"]["params"]["max_deposit_period"]="180s"' > $HOME/.seda-chain/config/tmp_genesis.json && mv $HOME/.seda-chain/config/tmp_genesis.json $HOME/.seda-chain/config/genesis.json

# TO-DO?
# - chain id
# - launch time

#
#   ADD GENESIS ACCOUNTS
#
$BIN add-genesis-account $ADDR1 100000000000000000seda --keyring-backend test
$BIN add-genesis-account $ADDR2 100000000000000000seda --keyring-backend test
$BIN add-genesis-account $ADDR3 100000000000000000seda --keyring-backend test
$BIN add-genesis-account $ADDR4 100000000000000000seda --keyring-backend test
$BIN add-genesis-account $ADDR5 100000000000000000seda --keyring-backend test
$BIN add-genesis-account $ADDR6 100000000000000000seda --keyring-backend test
$BIN add-genesis-account $ADDR7 100000000000000000seda --keyring-backend test
$BIN add-genesis-account $ADDR8 100000000000000000seda --keyring-backend test


#
# CREATE NODE KEY, VALIDATOR KEY, AND GENTX FOR EACH NODE
#
GENTX_DIR=$OUT_DIR/gentx
mkdir -p $GENTX_DIR

for i in ${!MONIKERS[@]}; do
    INDIVIDUAL_VAL_HOME_DIR=$OUT_DIR/${MONIKERS[$i]}
    INDIVIDUAL_VAL_CONFIG_DIR="$INDIVIDUAL_VAL_HOME_DIR/config"

    $BIN init new ${MONIKERS[$i]} --home $INDIVIDUAL_VAL_HOME_DIR
    $BIN keys add ${MONIKERS[$i]} --keyring-backend=test --home $INDIVIDUAL_VAL_HOME_DIR

    VALIDATOR_ADDRESS=$($BIN keys show ${MONIKERS[$i]} --keyring-backend test --home $INDIVIDUAL_VAL_HOME_DIR -a)

    # to create their gentx
    $BIN add-genesis-account $VALIDATOR_ADDRESS 100000000000000000seda --home $INDIVIDUAL_VAL_HOME_DIR
    # to output geneis file
    $BIN add-genesis-account $VALIDATOR_ADDRESS 100000000000000000seda

    $BIN gentx ${MONIKERS[$i]} ${SELF_DELEGATION_AMOUNTS[$i]} --moniker=${MONIKERS[$i]} --keyring-backend=test --home $INDIVIDUAL_VAL_HOME_DIR --ip=${IPS[$i]}

    cp -a $INDIVIDUAL_VAL_CONFIG_DIR/gentx/. $GENTX_DIR
done

cp -r $GENTX_DIR $CONFIG_DIR

$BIN collect-gentxs --home $HOME_DIR

cp $CONFIG_DIR/genesis.json $OUT_DIR
