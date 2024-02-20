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
if [ $($LOCAL_BIN version) != $CHAIN_VERSION ]; then
    echo "Local chain version is" $($LOCAL_BIN version) "instead of" $CHAIN_VERSION
    exit 1
fi

rm -rf $HOME_DIR
rm -rf $NODE_DIR

#
#   CREATE GENESIS AND ADJUST GENESIS PARAMETERS
#
$LOCAL_BIN init node0 --chain-id $CHAIN_ID --default-denom aseda

cat $HOME/.sedad/config/genesis.json | jq --arg GENESIS_TIME $GENESIS_TIME '.genesis_time=$GENESIS_TIME' > $HOME/.sedad/config/tmp_genesis.json && mv $HOME/.sedad/config/tmp_genesis.json $HOME/.sedad/config/genesis.json

# bank params
cat $HOME/.sedad/config/genesis.json | jq --argjson denom_metadata "$DENOM_METADATA" '.app_state["bank"]["denom_metadata"]=$denom_metadata' > $HOME/.sedad/config/tmp_genesis.json && mv $HOME/.sedad/config/tmp_genesis.json $HOME/.sedad/config/genesis.json

# crisis params
cat $HOME/.sedad/config/genesis.json | jq '.app_state["crisis"]["constant_fee"]["amount"]="1000000000000"' > $HOME/.sedad/config/tmp_genesis.json && mv $HOME/.sedad/config/tmp_genesis.json $HOME/.sedad/config/genesis.json

# distribution params
cat $HOME/.sedad/config/genesis.json | jq '.app_state["distribution"]["params"]["community_tax"]="0.000000000000000000"' > $HOME/.sedad/config/tmp_genesis.json && mv $HOME/.sedad/config/tmp_genesis.json $HOME/.sedad/config/genesis.json
cat $HOME/.sedad/config/genesis.json | jq '.app_state["distribution"]["params"]["base_proposer_reward"]="0.010000000000000000"' > $HOME/.sedad/config/tmp_genesis.json && mv $HOME/.sedad/config/tmp_genesis.json $HOME/.sedad/config/genesis.json
cat $HOME/.sedad/config/genesis.json | jq '.app_state["distribution"]["params"]["bonus_proposer_reward"]="0.040000000000000000"' > $HOME/.sedad/config/tmp_genesis.json && mv $HOME/.sedad/config/tmp_genesis.json $HOME/.sedad/config/genesis.json

# gov params
cat $HOME/.sedad/config/genesis.json | jq '.app_state["gov"]["voting_params"]["voting_period"]="432000s"' > $HOME/.sedad/config/tmp_genesis.json && mv $HOME/.sedad/config/tmp_genesis.json $HOME/.sedad/config/genesis.json
cat $HOME/.sedad/config/genesis.json | jq '.app_state["gov"]["params"]["voting_period"]="432000s"' > $HOME/.sedad/config/tmp_genesis.json && mv $HOME/.sedad/config/tmp_genesis.json $HOME/.sedad/config/genesis.json
# cat $HOME/.sedad/config/genesis.json | jq '.app_state["gov"]["params"]["expedited_voting_period"]="432000s"' > $HOME/.sedad/config/tmp_genesis.json && mv $HOME/.sedad/config/tmp_genesis.json $HOME/.sedad/config/genesis.json
cat $HOME/.sedad/config/genesis.json | jq '.app_state["gov"]["params"]["max_deposit_period"]="432000s"' > $HOME/.sedad/config/tmp_genesis.json && mv $HOME/.sedad/config/tmp_genesis.json $HOME/.sedad/config/genesis.json
cat $HOME/.sedad/config/genesis.json | jq '.app_state["gov"]["params"]["min_initial_deposit_ratio"]="0.010000000000000000"' > $HOME/.sedad/config/tmp_genesis.json && mv $HOME/.sedad/config/tmp_genesis.json $HOME/.sedad/config/genesis.json

# mint params
cat $HOME/.sedad/config/genesis.json | jq '.app_state["mint"]["params"]["blocks_per_year"]="4204800"' > $HOME/.sedad/config/tmp_genesis.json && mv $HOME/.sedad/config/tmp_genesis.json $HOME/.sedad/config/genesis.json

# slashing params
cat $HOME/.sedad/config/genesis.json | jq '.app_state["slashing"]["params"]["signed_blocks_window"]="10000"' > $HOME/.sedad/config/tmp_genesis.json && mv $HOME/.sedad/config/tmp_genesis.json $HOME/.sedad/config/genesis.json
cat $HOME/.sedad/config/genesis.json | jq '.app_state["slashing"]["params"]["min_signed_per_window"]="0.050000000000000000"' > $HOME/.sedad/config/tmp_genesis.json && mv $HOME/.sedad/config/tmp_genesis.json $HOME/.sedad/config/genesis.json
cat $HOME/.sedad/config/genesis.json | jq '.app_state["slashing"]["params"]["downtime_jail_duration"]="600s"' > $HOME/.sedad/config/tmp_genesis.json && mv $HOME/.sedad/config/tmp_genesis.json $HOME/.sedad/config/genesis.json
cat $HOME/.sedad/config/genesis.json | jq '.app_state["slashing"]["params"]["slash_fraction_double_sign"]="0.050000000000000000"' > $HOME/.sedad/config/tmp_genesis.json && mv $HOME/.sedad/config/tmp_genesis.json $HOME/.sedad/config/genesis.json
cat $HOME/.sedad/config/genesis.json | jq '.app_state["slashing"]["params"]["slash_fraction_downtime"]="0.0001"' > $HOME/.sedad/config/tmp_genesis.json && mv $HOME/.sedad/config/tmp_genesis.json $HOME/.sedad/config/genesis.json

# consensus params
cat $HOME/.sedad/config/genesis.json | jq '.consensus["params"]["block"]["max_gas"]="100000000"' > $HOME/.sedad/config/tmp_genesis.json && mv $HOME/.sedad/config/tmp_genesis.json $HOME/.sedad/config/genesis.json

# TO-DO wasm params

#
#   ADD GENESIS ACCOUNTS
#
for i in ${!GENESIS_ADDRESSES[@]}; do
    $LOCAL_BIN add-genesis-account ${GENESIS_ADDRESSES[$i]} 2000000seda --vesting-amount 1000000seda --vesting-start-time 1708610400 --vesting-end-time 1716386400 --funder $FUNDER_ADDRESS # 2M (1M nonvesting - 1M vesting)
done

set +u
if [ ! -z "$SATOSHI" ]; then
    $LOCAL_BIN add-genesis-account $SATOSHI 270000000seda # 270M
fi
if [ ! -z "$FAUCET" ]; then
    $LOCAL_BIN add-genesis-account $FAUCET 700000000seda # 700M
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

    $LOCAL_BIN init ${MONIKERS[$i]} --home $INDIVIDUAL_VAL_HOME_DIR  --chain-id $CHAIN_ID --default-denom aseda
    $LOCAL_BIN keys add ${MONIKERS[$i]} --keyring-backend=test --home $INDIVIDUAL_VAL_HOME_DIR

    VALIDATOR_ADDRESS=$($LOCAL_BIN keys show ${MONIKERS[$i]} --keyring-backend test --home $INDIVIDUAL_VAL_HOME_DIR -a)

    if [ -z ${VESTING_AMOUNTS[$i]} ]; then
        # to create their gentx
        $LOCAL_BIN add-genesis-account $VALIDATOR_ADDRESS ${SELF_DELEGATION_AMOUNTS[$i]} --home $INDIVIDUAL_VAL_HOME_DIR
        # to output geneis file
        $LOCAL_BIN add-genesis-account $VALIDATOR_ADDRESS ${SELF_DELEGATION_AMOUNTS[$i]}
    else
        # to create their gentx
        $LOCAL_BIN add-genesis-account $VALIDATOR_ADDRESS ${SELF_DELEGATION_AMOUNTS[$i]} --home $INDIVIDUAL_VAL_HOME_DIR --vesting-amount ${VESTING_AMOUNTS[$i]} --vesting-start-time 1708610400 --vesting-end-time 1716386400 --funder $FUNDER_ADDRESS
        # to output geneis file
        $LOCAL_BIN add-genesis-account $VALIDATOR_ADDRESS ${SELF_DELEGATION_AMOUNTS[$i]} --vesting-amount ${VESTING_AMOUNTS[$i]} --vesting-start-time 1708610400 --vesting-end-time 1716386400 --funder $FUNDER_ADDRESS
    fi


    $LOCAL_BIN gentx ${MONIKERS[$i]} ${SELF_DELEGATION_AMOUNTS[$i]} --moniker=${MONIKERS[$i]} --keyring-backend=test --home $INDIVIDUAL_VAL_HOME_DIR --ip=${IPS[$i]} --chain-id $CHAIN_ID

    cp -a $INDIVIDUAL_VAL_CONFIG_DIR/gentx/. $GENTX_DIR
done

cp -r $GENTX_DIR $HOME_CONFIG_DIR
$LOCAL_BIN collect-gentxs --home $HOME_DIR
cp $HOME_CONFIG_DIR/genesis.json $NODE_DIR
