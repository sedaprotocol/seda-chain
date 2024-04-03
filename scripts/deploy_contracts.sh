#!/bin/bash
set -e
set -x

CHAIN_ID=seda-1-local
RPC_URL=http://127.0.0.1:26657
BIN=$(git rev-parse --show-toplevel)/build/sedad
ARTIFACTS_DIR=$(git rev-parse --show-toplevel)/testutil/testwasms # must contain proxy_contract.wasm, data_requests.wasm, and staking.wasm
VOTING_PERIOD=30 # seconds
DEV_ACCOUNT=$($BIN keys show satoshi --keyring-backend test -a) # for sending wasm-storage txs
VOTE_ACCOUNT=$($BIN keys show satoshi --keyring-backend test -a) # for sending vote txs


echo "Deploying proxy contract"

# Proxy Contracts
OUTPUT="$($BIN tx wasm store $ARTIFACTS_DIR/proxy_contract.wasm --node $RPC_URL --from $DEV_ACCOUNT --keyring-backend test --gas-prices 100000000000aseda --gas auto --gas-adjustment 1.3 -y --output json --chain-id $CHAIN_ID)"
TXHASH=$(echo $OUTPUT | jq -r '.txhash')

sleep 10

OUTPUT="$($BIN query tx $TXHASH --node $RPC_URL --output json)"
PROXY_CODE_ID=$(echo "$OUTPUT" | jq -r '.events[] | select(.type=="store_code") | .attributes[] | select(.key=="code_id") | .value')

echo "Instantiating proxy contract on code id $PROXY_CODE_ID"

OUTPUT=$($BIN tx wasm-storage submit-proposal instantiate-and-register-proxy-contract $PROXY_CODE_ID \
    '{"token":"aseda"}' 74657374696e67 --admin $DEV_ACCOUNT \
    --label proxy$PROXY_CODE_ID \
    --title 'Proxy Contract' --summary 'Instantiates and registers proxy contract' --deposit 10000000aseda \
    --from $DEV_ACCOUNT --keyring-backend test \
    --node $RPC_URL \
    --gas-prices 100000000000aseda --gas auto --gas-adjustment 1.3 \
    --output json --chain-id $CHAIN_ID -y)
TXHASH=$(echo "$OUTPUT" | jq -r '.txhash')

sleep 10

PROPOSAL_ID=$($BIN query tx $TXHASH --output json | jq '.events[] | select(.type == "submit_proposal") | .attributes[] | select(.key == "proposal_id") | .value' | sed 's/^"\(.*\)"$/\1/')  
$BIN tx gov vote $PROPOSAL_ID yes \
    --from $VOTE_ACCOUNT --keyring-backend test \
    --gas-prices 100000000000aseda --gas auto --gas-adjustment 1.6 \
    --chain-id $CHAIN_ID -y

sleep $VOTING_PERIOD

PROXY_CONTRACT_ADDRESS=$($BIN query wasm-storage proxy-contract-registry --output json | jq -r '.address')
echo "Deployed proxy contract to: $PROXY_CONTRACT_ADDRESS"

# -----------------------
# Data Requests
# -----------------------

echo "Deploying Data Request contract"

OUTPUT="$($BIN tx wasm store $ARTIFACTS_DIR/data_requests.wasm --node $RPC_URL --from $DEV_ACCOUNT --keyring-backend test --gas-prices 100000000000aseda --gas auto --gas-adjustment 1.3 -y --output json --chain-id $CHAIN_ID)"
TXHASH=$(echo $OUTPUT | jq -r '.txhash')
sleep 10
OUTPUT="$($BIN query tx $TXHASH --node $RPC_URL --output json)"
DRs_CODE_ID=$(echo "$OUTPUT" | jq -r '.events[] | select(.type=="store_code") | .attributes[] | select(.key=="code_id") | .value')

echo "Instantiating data request contract on code id $DRs_CODE_ID"

OUTPUT=$($BIN tx wasm instantiate $DRs_CODE_ID '{"token":"aseda", "proxy": "'$PROXY_CONTRACT_ADDRESS'" }' --no-admin --from $DEV_ACCOUNT --keyring-backend test --node $RPC_URL --label dr$DRs_CODE_ID --gas-prices 100000000000aseda --gas auto --gas-adjustment 1.3 -y --output json --chain-id $CHAIN_ID)
TXHASH=$(echo "$OUTPUT" | jq -r '.txhash')
sleep 10
OUTPUT="$($BIN query tx $TXHASH --node $RPC_URL --output json)"
DRs_CONTRACT_ADDRESS=$(echo "$OUTPUT" | jq -r '.events[] | select(.type=="instantiate") | .attributes[] | select(.key=="_contract_address") | .value')

echo "Deployed data request contract to: $DRs_CONTRACT_ADDRESS"

# -----------------------
# Staking
# -----------------------

echo "Deploying staking contract"

OUTPUT="$($BIN tx wasm store $ARTIFACTS_DIR/staking.wasm --node $RPC_URL --from $DEV_ACCOUNT --keyring-backend test --gas-prices 100000000000aseda --gas auto --gas-adjustment 1.3 -y --output json --chain-id $CHAIN_ID)"
TXHASH=$(echo $OUTPUT | jq -r '.txhash')
sleep 10
OUTPUT="$($BIN query tx $TXHASH --node $RPC_URL --output json)"
STAKING_CODE_ID=$(echo "$OUTPUT" | jq -r '.events[] | select(.type=="store_code") | .attributes[] | select(.key=="code_id") | .value')

echo "Instantiating staking contract on code id $STAKING_CODE_ID"

OUTPUT=$($BIN tx wasm instantiate $STAKING_CODE_ID '{"token":"aseda", "proxy":  "'$PROXY_CONTRACT_ADDRESS'" }' --no-admin --from $DEV_ACCOUNT --keyring-backend test --node $RPC_URL --label staking$staking_code_id --gas-prices 100000000000aseda --gas auto --gas-adjustment 1.3 -y --output json --chain-id $CHAIN_ID)
TXHASH=$(echo "$OUTPUT" | jq -r '.txhash')
sleep 10
OUTPUT="$($BIN query tx $TXHASH --node $RPC_URL --output json)"
STAKING_CONTRACT_ADDRESS=$(echo "$OUTPUT" | jq -r '.events[] | select(.type=="instantiate") | .attributes[] | select(.key=="_contract_address") | .value')

echo "Deployed staking contract to: $STAKING_CONTRACT_ADDRESS"

# -----------------------
# Setting properties
# -----------------------

echo "Setting dr address on proxy.."
OUTPUT="$($BIN tx wasm execute $PROXY_CONTRACT_ADDRESS '{"set_data_requests":{"contract": "'$DRs_CONTRACT_ADDRESS'" }}' --from $DEV_ACCOUNT --keyring-backend test --node $RPC_URL --gas-prices 100000000000aseda --gas auto --gas-adjustment 1.3 -y --output json --chain-id $CHAIN_ID)"
TXHASH=$(echo "$OUTPUT" | jq -r '.txhash')
sleep 10
OUTPUT="$($BIN query tx $TXHASH --node $RPC_URL --output json)"
echo "$OUTPUT"

echo "Setting staking address on proxy.."
OUTPUT="$($BIN tx wasm execute $PROXY_CONTRACT_ADDRESS '{"set_staking":{"contract": "'$STAKING_CONTRACT_ADDRESS'" }}' --from $DEV_ACCOUNT --keyring-backend test --node $RPC_URL --gas-prices 100000000000aseda --gas auto --gas-adjustment 1.3 -y --output json --chain-id $CHAIN_ID)"
TXHASH=$(echo "$OUTPUT" | jq -r '.txhash')
sleep 10
OUTPUT="$($BIN query tx $TXHASH --node $RPC_URL --output json)"
echo "$OUTPUT"
