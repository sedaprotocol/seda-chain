#!/bin/bash
set -euxo pipefail

#
# This script dumps wasm states after storing and instantiating contracts.
# Then, it adds these states to a given original genesis file.
# The final genesis file is placed in the current directory as genesis.json.
#

#
#   PARAMETERS
#
WASM_DIR=./artifacts # where Wasm files are located
BIN=../../build/seda-chaind # chain binary executable on your machine
ORIGINAL_GENESIS=./nodes/genesis.json # genesis file to be modified by this script


#
#   PRELIMINARY CHECKS
#
if [ ! -f "$ORIGINAL_GENESIS" ]; then
  echo "Original genesis file not found."
  exit 1
fi

TMP_HOME=./tmp
rm -rf $TMP_HOME


# store_and_instantiate() stores and instantiates a contract and returns its address
# Accept arguments:
#   $1: Contract file name
#   $2: Initial state
store_and_instantiate() {
    local TX_OUTPUT=$($BIN tx wasm store $WASM_DIR/$1 --from $ADDR --keyring-backend test --gas auto --gas-adjustment 1.2 --home $TMP_HOME -y --output json)
    [[ -z "$TX_OUTPUT" ]] && { echo "failed to get tx output" ; exit 1; }
    local TX_HASH=$(echo $TX_OUTPUT | jq -r .txhash)
    sleep 10;

    local STORE_TX_OUTPUT=$($BIN query tx $TX_HASH  --home $TMP_HOME --output json)
    local CODE_ID=$(echo $STORE_TX_OUTPUT | jq -r '.events[] | select(.type | contains("store_code")).attributes[] | select(.key | contains("code_id")).value')
    [[ -z "$CODE_ID" ]] && { echo "failed to get code ID" ; exit 1; }

    local INSTANTIATE_OUTPUT=$($BIN tx wasm instantiate $CODE_ID "$2" --no-admin --from $ADDR --keyring-backend test --label $CODE_ID --gas auto --gas-adjustment 1.2 --home $TMP_HOME -y --output json)
    TX_HASH=$(echo "$INSTANTIATE_OUTPUT" | jq -r '.txhash')
    sleep 10;

    local INSTANTIATE_TX_OUTPUT=$($BIN query tx $TX_HASH  --home $TMP_HOME --output json)
    local CONTRACT_ADDRESS=$(echo $INSTANTIATE_TX_OUTPUT | jq -r '.logs[].events[] | select(.type=="instantiate") | .attributes[] | select(.key=="_contract_address") | .value')
    [[ -z "$CONTRACT_ADDRESS" ]] && { echo "failed to get contract address for ${1}" ; exit 1; }
    
    echo $CONTRACT_ADDRESS
}


#
#   SCRIPT BEGINS - START CHAIN
#
$BIN init new node0 --home $TMP_HOME

$BIN keys add satoshi --home $TMP_HOME --keyring-backend test
ADDR=$($BIN keys show satoshi --home $TMP_HOME --keyring-backend test -a)
$BIN add-genesis-account $ADDR 100000000000000000seda --home $TMP_HOME --keyring-backend test
$BIN gentx satoshi 10000000000000000seda --home $TMP_HOME --keyring-backend test
$BIN collect-gentxs --home $TMP_HOME


$BIN start --home $TMP_HOME > chain_output.log 2>&1 & disown

sleep 20


#
#   SEND TRANSACTIONS WHILE CHAIN IS RUNNING
#

# Store and instantiate three contracts
PROXY_ADDR=$(store_and_instantiate proxy_contract.wasm '{"token":"aseda"}')

ARG='{"token":"aseda", "proxy": "'$PROXY_ADDR'" }'
STAKING_ADDR=$(store_and_instantiate staking.wasm "$ARG")
DR_ADDR=$(store_and_instantiate data_requests.wasm "$ARG")


# Call SetStaking and SetDataRequests on Proxy contract to set circular dependency
$BIN tx wasm execute $PROXY_ADDR '{"set_staking":{"contract":"'$STAKING_ADDR'"}}' --from $ADDR --gas auto --gas-adjustment 1.2 --keyring-backend test  --home $TMP_HOME -y
sleep 10
$BIN tx wasm execute $PROXY_ADDR '{"set_data_requests":{"contract":"'$DR_ADDR'"}}' --from $ADDR --gas auto --gas-adjustment 1.2 --keyring-backend test  --home $TMP_HOME -y
sleep 10



#
#   TERMINATE CHAIN PROCESS, EXPORT, AND MODIFY GIVEN GENESIS
#
pkill seda-chaind


$BIN export --home $TMP_HOME > $TMP_HOME/exported
python3 -m json.tool $TMP_HOME/exported > $TMP_HOME/genesis.json
rm $TMP_HOME/exported


EXPORTED_GENESIS=$TMP_HOME/genesis.json
TMP_GENESIS=$TMP_HOME/tmp_genesis.json
TMP_TMP_GENESIS=$TMP_HOME/tmp_tmp_genesis.json

#
# Modify
# - wasm.codes
# - wasm.contracts
# - wasm.sequences
# - wasm-storage.proxy_contract_registry
#
CODES=$(jq '.app_state["wasm"]["codes"]' $EXPORTED_GENESIS)
CONTRACTS=$(jq '.app_state["wasm"]["contracts"]' $EXPORTED_GENESIS)
SEQUENCES=$(jq '.app_state["wasm"]["sequences"]' $EXPORTED_GENESIS)

jq '.app_state["wasm-storage"]["proxy_contract_registry"]="'$PROXY_ADDR'"' "$ORIGINAL_GENESIS" > "$TMP_TMP_GENESIS" && mv $TMP_TMP_GENESIS $TMP_GENESIS
jq '.app_state["wasm"]["codes"]='"$CODES"'' "$TMP_GENESIS" > "$TMP_TMP_GENESIS" && mv $TMP_TMP_GENESIS $TMP_GENESIS
jq '.app_state["wasm"]["contracts"]='"$CONTRACTS"'' "$TMP_GENESIS" > "$TMP_TMP_GENESIS" && mv $TMP_TMP_GENESIS $TMP_GENESIS
jq '.app_state["wasm"]["sequences"]='"$SEQUENCES"'' "$TMP_GENESIS" > "$TMP_TMP_GENESIS" && mv $TMP_TMP_GENESIS $TMP_GENESIS

mv $TMP_GENESIS $ORIGINAL_GENESIS
rm -rf $TMP_HOME
