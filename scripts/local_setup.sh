#!/bin/bash
set -e
set -x

BIN=./build/sedad
CONFIG_PATH=$HOME/.sedad/config

function add_key_and_account() {
    local name=$1
    local amount=$2
    if [ -n "$3" ]; then
        echo $3 | $BIN keys add $name --keyring-backend test --recover
    else
        $BIN keys add $name --keyring-backend test
    fi
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
rm -rf ~/.sedad || true

# configure sedad
$BIN config set client chain-id seda-1-local

# initialize the chain
$BIN init node0 --default-denom aseda

cat $HOME/.sedad/config/genesis.json | jq '.app_state["gov"]["voting_params"]["voting_period"]="30s"' > $HOME/.sedad/config/tmp_genesis.json && mv $HOME/.sedad/config/tmp_genesis.json $HOME/.sedad/config/genesis.json
cat $HOME/.sedad/config/genesis.json | jq '.app_state["gov"]["params"]["voting_period"]="30s"' > $HOME/.sedad/config/tmp_genesis.json && mv $HOME/.sedad/config/tmp_genesis.json $HOME/.sedad/config/genesis.json
cat $HOME/.sedad/config/genesis.json | jq '.app_state["gov"]["params"]["expedited_voting_period"]="15s"' > $HOME/.sedad/config/tmp_genesis.json && mv $HOME/.sedad/config/tmp_genesis.json $HOME/.sedad/config/genesis.json
cat $HOME/.sedad/config/genesis.json | jq '.consensus_params["block"]["max_gas"]="100000000"' > $HOME/.sedad/config/tmp_genesis.json && mv $HOME/.sedad/config/tmp_genesis.json $HOME/.sedad/config/genesis.json

if [[ "$OSTYPE" == "darwin"* ]]; then
    sed -i '' 's/swagger = false/swagger = true/' $HOME/.sedad/config/app.toml
else
    sed 's/swagger = false/swagger = true/' $HOME/.sedad/config/app.toml > tmp
    cat tmp > $HOME/.sedad/config/app.toml 
    rm tmp
fi

# update genesis
add_key_and_account "fixedacc" "100000000000000000seda" "tortoise chunk claim human keen potato venue follow physical weasel famous series source upgrade give rare gossip practice artist truly shell buddy garment design"
add_key_and_account "fixedacc2" "100000000000000000seda" "hole bag crumble table stage eternal gather two cabbage define write update run biology side deal great casual absorb panther month better heart trigger"
add_key_and_account "satoshi" "100000000000000000seda"
add_key_and_account "acc1" "100000000000000000seda"

$BIN keys add vesttest --keyring-backend test
VESTING_START=$(date +%s)
VESTING_END=$((VESTING_START + 100000))
$BIN add-genesis-account vesttest 10000seda --vesting-amount 10000seda --vesting-start-time $VESTING_START --vesting-end-time $VESTING_END --funder seda1jq60my60e87arglrazfpqn753hx0pzcatdek76 --keyring-backend test

# create a default validator
$BIN gentx satoshi 10000000000000000seda --keyring-backend test

# collect genesis txns
$BIN collect-gentxs

# start the chain
$BIN start --log_level debug 2>&1 | tee local_chain.log
