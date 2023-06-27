DEPLOYER_ADDRESS=cosmos1njge6f3h8huhvnd72whzgyckgdp64q5xnus7k9
CONTRACT_PATH=./scripts/seda_chain_contracts-aarch64.wasm
INITIAL_STATE="{}"
SHOULD_POST_DR=true

CONTRACT_DEPLOY_OUTPUT=$(seda-chaind tx wasm store $CONTRACT_PATH --from $DEPLOYER_ADDRESS --gas-prices 0.1seda --gas auto --gas-adjustment 1.3 -y --output json)
CONTRACT_DEPLOY_TX_HASH=$(echo $CONTRACT_DEPLOY_OUTPUT | jq -r .txhash)

sleep 2;

CONTRACT_DEPLOY_TX_OUTPUT=$(seda-chaind query tx $CONTRACT_DEPLOY_TX_HASH --output json)
CONTRACT_DEPLOY_CODE_ID=$(echo $CONTRACT_DEPLOY_TX_OUTPUT | jq -r '.events[] | select(.type | contains("store_code")).attributes[] | select(.key | contains("code_id")).value')

INSTANTIATE_OUTPUT=$(seda-chaind tx wasm instantiate $CONTRACT_DEPLOY_CODE_ID $INITIAL_STATE --admin="$(seda-chaind keys show $DEPLOYER_ADDRESS -a)" --from $DEPLOYER_ADDRESS --label "local0.1.0" -y --output json) 
INSTANTIATE_TX_HASH=$(echo $INSTANTIATE_OUTPUT | jq -r .txhash)

sleep 2;

INSTANTIATE_TX_OUTPUT=$(seda-chaind query tx $INSTANTIATE_TX_HASH --output json)
CONTRACT_ADDRESS=$(echo $INSTANTIATE_TX_OUTPUT | jq -r '.events[] | select(.type | contains("instantiate")).attributes[] | select(.key | contains("_contract_address")).value')

echo $CONTRACT_ADDRESS;