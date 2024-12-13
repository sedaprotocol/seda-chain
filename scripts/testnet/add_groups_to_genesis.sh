#!/bin/bash
set -euxo pipefail

source config.sh

########################################
#   PRELIMINARY CHECKS AND DOWNLOADS   #
########################################
ORIGINAL_GENESIS=$NODE_DIR/genesis.json
if [ ! -f "$ORIGINAL_GENESIS" ]; then
  echo "Original genesis file not found inside node directory."
  exit 1
fi

TMP_HOME=./tmp
rm -rf $TMP_HOME

TEMP_CHAIN_ID=temp-seda-chain

###################################
#   SCRIPT BEGINS - START CHAIN   #
###################################
$LOCAL_BIN init node0 --home $TMP_HOME --chain-id $TEMP_CHAIN_ID --default-denom aseda

$LOCAL_BIN keys add deployer --home $TMP_HOME --keyring-backend test
ADDR=$($LOCAL_BIN keys show deployer --home $TMP_HOME --keyring-backend test -a)
$LOCAL_BIN add-genesis-account $ADDR 100000000000000000seda --home $TMP_HOME --keyring-backend test

echo $ADMIN_SEED | $LOCAL_BIN keys add admin --home $TMP_HOME --keyring-backend test --recover
ADMIN_ADDR=$($LOCAL_BIN keys show admin --home $TMP_HOME --keyring-backend test -a)
$LOCAL_BIN add-genesis-account $ADMIN_ADDR 100000000000000000seda --home $TMP_HOME --keyring-backend test

$LOCAL_BIN gentx deployer 10000000000000000seda --home $TMP_HOME --keyring-backend test --chain-id $TEMP_CHAIN_ID
$LOCAL_BIN collect-gentxs --home $TMP_HOME

$LOCAL_BIN start --home $TMP_HOME > chain_output.log 2>&1 & disown

sleep 20


###############################################
#  SEND TRANSACTIONS WHILE CHAIN IS RUNNING   #
###############################################

# Create group and group policy
# SEDA Security Policy
$LOCAL_BIN tx group create-group-with-policy $ADMIN_ADDR "Security Group" "{\"name\":\"Security Group Policy\",\"description\":\"\"}" $GROUP_SECURITY_MEMBERS $GROUP_SECURITY_POLICY --home $TMP_HOME --from $ADMIN_ADDR --keyring-backend test --fees 1seda --gas auto --gas-adjustment 1.5 --chain-id $TEMP_CHAIN_ID -y
sleep 10
# DAO Treasury Group
$LOCAL_BIN tx group create-group-with-policy $ADMIN_ADDR "Treasury Group" "{\"name\":\"Treasury Group Policy\",\"description\":\"\"}" $GROUP_TREASURY_MEMBERS $GROUP_TREASURY_POLICY --home $TMP_HOME --from $ADMIN_ADDR --keyring-backend test --fees 1seda --gas auto --gas-adjustment 1.5 --chain-id $TEMP_CHAIN_ID -y
sleep 10
# OOA Group
$LOCAL_BIN tx group create-group-with-policy $ADMIN_ADDR "OOA Group" "{\"name\":\"OOA Group Policy\",\"description\":\"\"}" $GROUP_OOA_MEMBERS $GROUP_OOA_POLICY --home $TMP_HOME --from $ADMIN_ADDR --keyring-backend test --fees 1seda --gas auto --gas-adjustment 1.5 --chain-id $TEMP_CHAIN_ID -y
sleep 10


#############################################################
# TERMINATE CHAIN PROCESS, EXPORT, AND MODIFY GIVEN GENESIS #
#############################################################
pkill sedad
sleep 5

$LOCAL_BIN export --home $TMP_HOME > $TMP_HOME/exported
python3 -m json.tool $TMP_HOME/exported > $TMP_HOME/genesis.json
rm $TMP_HOME/exported


EXPORTED_GENESIS=$TMP_HOME/genesis.json
TMP_GENESIS=$TMP_HOME/tmp_genesis.json
TMP_TMP_GENESIS=$TMP_HOME/tmp_tmp_genesis.json

cp $ORIGINAL_GENESIS $TMP_GENESIS # make adjustments on TMP_GENESIS until replacing original genesis in the last step

# Modify group state and wasm code upload params
jq '.app_state["group"]' $EXPORTED_GENESIS > $TMP_HOME/group.tmp
TREASURY_GROUP_POLICY_ADDR=$(jq '.app_state["group"]["group_policies"][0]["address"]' $EXPORTED_GENESIS)
OOA_GROUP_POLICY_ADDR=$(jq '.app_state["group"]["group_policies"][1]["address"]' $EXPORTED_GENESIS)
SECURITY_GROUP_POLICY_ADDR=$(jq '.app_state["group"]["group_policies"][2]["address"]' $EXPORTED_GENESIS)

jq --slurpfile group $TMP_HOME/group.tmp '.app_state["group"] = $group[0]' $TMP_GENESIS > $TMP_TMP_GENESIS && mv $TMP_TMP_GENESIS $TMP_GENESIS

# replace group policy address as group & group policy admin
jq '(.app_state.group.groups[] | select(.metadata == "Security Group") .admin) |= '$SECURITY_GROUP_POLICY_ADDR'' $TMP_GENESIS > $TMP_TMP_GENESIS && mv $TMP_TMP_GENESIS $TMP_GENESIS
jq '(.app_state.group.group_policies[] | select(.address == '$SECURITY_GROUP_POLICY_ADDR') .admin) |= '$SECURITY_GROUP_POLICY_ADDR'' $TMP_GENESIS > $TMP_TMP_GENESIS && mv $TMP_TMP_GENESIS $TMP_GENESIS

jq '(.app_state.group.groups[] | select(.metadata == "Treasury Group") .admin) |= '$TREASURY_GROUP_POLICY_ADDR'' $TMP_GENESIS > $TMP_TMP_GENESIS && mv $TMP_TMP_GENESIS $TMP_GENESIS
jq '(.app_state.group.group_policies[] | select(.address == '$TREASURY_GROUP_POLICY_ADDR') .admin) |= '$TREASURY_GROUP_POLICY_ADDR'' $TMP_GENESIS > $TMP_TMP_GENESIS && mv $TMP_TMP_GENESIS $TMP_GENESIS

jq '(.app_state.group.groups[] | select(.metadata == "OOA Group") .admin) |= '$OOA_GROUP_POLICY_ADDR'' $TMP_GENESIS > $TMP_TMP_GENESIS && mv $TMP_TMP_GENESIS $TMP_GENESIS
jq '(.app_state.group.group_policies[] | select(.address == '$OOA_GROUP_POLICY_ADDR') .admin) |= '$OOA_GROUP_POLICY_ADDR'' $TMP_GENESIS > $TMP_TMP_GENESIS && mv $TMP_TMP_GENESIS $TMP_GENESIS

if [ "$WASM_PERMISSION_EVERYONE" != "true" ]; then
  jq '.app_state["wasm"]["params"]["code_upload_access"]["permission"]="AnyOfAddresses"' $TMP_GENESIS > $TMP_TMP_GENESIS && mv $TMP_TMP_GENESIS $TMP_GENESIS
  jq '.app_state["wasm"]["params"]["instantiate_default_permission"]="AnyOfAddresses"' $TMP_GENESIS > $TMP_TMP_GENESIS && mv $TMP_TMP_GENESIS $TMP_GENESIS
  jq '.app_state["wasm"]["params"]["code_upload_access"]["addresses"]=['$SECURITY_GROUP_POLICY_ADDR']' $TMP_GENESIS > $TMP_TMP_GENESIS && mv $TMP_TMP_GENESIS $TMP_GENESIS
fi

mv $TMP_GENESIS $ORIGINAL_GENESIS

# clean up
rm -rf $TMP_HOME
rm chain_output.log
