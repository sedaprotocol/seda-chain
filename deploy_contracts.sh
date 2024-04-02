CHAIN_ID=seda-1-local
RPC_URL=http://127.0.0.1:26657
BIN=./build/sedad

DEV_ACCOUNT=$($BIN keys show franklin3 --keyring-backend test -a)

echo "Deploying proxy contract"

# Proxy Contracts
OUTPUT="$($BIN tx wasm store ./artifacts/proxy_contract.wasm --node $RPC_URL --from $DEV_ACCOUNT --keyring-backend test --gas-prices 0.1aseda --gas auto --gas-adjustment 1.3 -y --output json --chain-id $CHAIN_ID)"
TXHASH=$(echo $OUTPUT | jq -r '.txhash')

sleep 10

OUTPUT="$($BIN query tx $TXHASH --node $RPC_URL --output json)"
PROXY_CODE_ID=$(echo "$OUTPUT" | jq -r '.events[] | select(.type=="store_code") | .attributes[] | select(.key=="code_id") | .value')

echo "Instantiating proxy contract on code id $PROXY_CODE_ID"

OUTPUT=$($BIN tx wasm instantiate $PROXY_CODE_ID '{"token":"aseda"}' --no-admin --from $DEV_ACCOUNT --keyring-backend test --node $RPC_URL --label proxy$PROXY_CODE_ID --gas-prices 0.1aseda --gas auto --gas-adjustment 1.3 -y --output json --chain-id $CHAIN_ID)
TXHASH=$(echo "$OUTPUT" | jq -r '.txhash')

sleep 10

OUTPUT="$($BIN query tx $TXHASH --node $RPC_URL --output json)"
PROXY_CONTRACT_ADDRESS=$(echo "$OUTPUT" | jq -r '.events[] | select(.type=="instantiate") | .attributes[] | select(.key=="_contract_address") | .value')

echo "Deployed proxy contract to: $PROXY_CONTRACT_ADDRESS"

# -----------------------
# Data Requests
# -----------------------

echo "Deploying Data Request contract"

OUTPUT="$($BIN tx wasm store artifacts/data_requests.wasm --node $RPC_URL --from $DEV_ACCOUNT --keyring-backend test --gas-prices 0.1aseda --gas auto --gas-adjustment 1.3 -y --output json --chain-id $CHAIN_ID)"
TXHASH=$(echo $OUTPUT | jq -r '.txhash')
sleep 10
OUTPUT="$($BIN query tx $TXHASH --node $RPC_URL --output json)"
DRs_CODE_ID=$(echo "$OUTPUT" | jq -r '.events[] | select(.type=="store_code") | .attributes[] | select(.key=="code_id") | .value')

echo "Instantiating data request contract on code id $DRs_CODE_ID"

OUTPUT=$($BIN tx wasm instantiate $DRs_CODE_ID '{"token":"aseda", "proxy": "'$PROXY_CONTRACT_ADDRESS'" }' --no-admin --from $DEV_ACCOUNT --keyring-backend test --node $RPC_URL --label dr$DRs_CODE_ID --gas-prices 0.1aseda --gas auto --gas-adjustment 1.3 -y --output json --chain-id $CHAIN_ID)
TXHASH=$(echo "$OUTPUT" | jq -r '.txhash')
sleep 10
OUTPUT="$($BIN query tx $TXHASH --node $RPC_URL --output json)"
DRs_CONTRACT_ADDRESS=$(echo "$OUTPUT" | jq -r '.events[] | select(.type=="instantiate") | .attributes[] | select(.key=="_contract_address") | .value')

echo "Deployed data request contract to: $DRs_CONTRACT_ADDRESS"

# -----------------------
# Staking
# -----------------------

echo "Deploying staking contract"

OUTPUT="$($BIN tx wasm store artifacts/staking.wasm --node $RPC_URL --from $DEV_ACCOUNT --keyring-backend test --gas-prices 0.1aseda --gas auto --gas-adjustment 1.3 -y --output json --chain-id $CHAIN_ID)"
TXHASH=$(echo $OUTPUT | jq -r '.txhash')
sleep 10
OUTPUT="$($BIN query tx $TXHASH --node $RPC_URL --output json)"
STAKING_CODE_ID=$(echo "$OUTPUT" | jq -r '.events[] | select(.type=="store_code") | .attributes[] | select(.key=="code_id") | .value')

echo "Instantiating staking contract on code id $STAKING_CODE_ID"

OUTPUT=$($BIN tx wasm instantiate $STAKING_CODE_ID '{"token":"aseda", "proxy":  "'$PROXY_CONTRACT_ADDRESS'" }' --no-admin --from $DEV_ACCOUNT --keyring-backend test --node $RPC_URL --label staking$staking_code_id --gas-prices 0.1aseda --gas auto --gas-adjustment 1.3 -y --output json --chain-id $CHAIN_ID)
TXHASH=$(echo "$OUTPUT" | jq -r '.txhash')
sleep 10
OUTPUT="$($BIN query tx $TXHASH --node $RPC_URL --output json)"
STAKING_CONTRACT_ADDRESS=$(echo "$OUTPUT" | jq -r '.events[] | select(.type=="instantiate") | .attributes[] | select(.key=="_contract_address") | .value')

echo "Deployed staking contract to: $STAKING_CONTRACT_ADDRESS"

# -----------------------
# Setting properties
# -----------------------

echo "Setting dr address on proxy.."
OUTPUT="$($BIN tx wasm execute $PROXY_CONTRACT_ADDRESS '{"set_data_requests":{"contract": "'$DRs_CONTRACT_ADDRESS'" }}' --from $DEV_ACCOUNT --keyring-backend test --node $RPC_URL --gas-prices 0.1aseda --gas auto --gas-adjustment 1.3 -y --output json --chain-id $CHAIN_ID)"
TXHASH=$(echo "$OUTPUT" | jq -r '.txhash')
sleep 10
OUTPUT="$($BIN query tx $TXHASH --node $RPC_URL --output json)"
echo "$OUTPUT"

echo "Setting staking address on proxy.."
OUTPUT="$($BIN tx wasm execute $PROXY_CONTRACT_ADDRESS '{"set_staking":{"contract": "'$STAKING_CONTRACT_ADDRESS'" }}' --from $DEV_ACCOUNT --keyring-backend test --node $RPC_URL --gas-prices 0.1aseda --gas auto --gas-adjustment 1.3 -y --output json --chain-id $CHAIN_ID)"
TXHASH=$(echo "$OUTPUT" | jq -r '.txhash')
sleep 10
OUTPUT="$($BIN query tx $TXHASH --node $RPC_URL --output json)"
echo "$OUTPUT"