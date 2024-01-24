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

cat $HOME/.seda-chain/config/genesis.json | jq --arg GENESIS_TIME $GENESIS_TIME '.genesis_time=$GENESIS_TIME' > $HOME/.seda-chain/config/tmp_genesis.json && mv $HOME/.seda-chain/config/tmp_genesis.json $HOME/.seda-chain/config/genesis.json

# bank params
cat $HOME/.seda-chain/config/genesis.json | jq --argjson denom_metadata "$DENOM_METADATA" '.app_state["bank"]["denom_metadata"]=$denom_metadata' > $HOME/.seda-chain/config/tmp_genesis.json && mv $HOME/.seda-chain/config/tmp_genesis.json $HOME/.seda-chain/config/genesis.json

# crisis params
cat $HOME/.seda-chain/config/genesis.json | jq '.app_state["crisis"]["constant_fee"]["amount"]="1000000000000"' > $HOME/.seda-chain/config/tmp_genesis.json && mv $HOME/.seda-chain/config/tmp_genesis.json $HOME/.seda-chain/config/genesis.json

# distribution params
cat $HOME/.seda-chain/config/genesis.json | jq '.app_state["distribution"]["params"]["community_tax"]="0.000000000000000000"' > $HOME/.seda-chain/config/tmp_genesis.json && mv $HOME/.seda-chain/config/tmp_genesis.json $HOME/.seda-chain/config/genesis.json
cat $HOME/.seda-chain/config/genesis.json | jq '.app_state["distribution"]["params"]["base_proposer_reward"]="0.010000000000000000"' > $HOME/.seda-chain/config/tmp_genesis.json && mv $HOME/.seda-chain/config/tmp_genesis.json $HOME/.seda-chain/config/genesis.json
cat $HOME/.seda-chain/config/genesis.json | jq '.app_state["distribution"]["params"]["bonus_proposer_reward"]="0.040000000000000000"' > $HOME/.seda-chain/config/tmp_genesis.json && mv $HOME/.seda-chain/config/tmp_genesis.json $HOME/.seda-chain/config/genesis.json

# gov params
cat $HOME/.seda-chain/config/genesis.json | jq '.app_state["gov"]["voting_params"]["voting_period"]="180s"' > $HOME/.seda-chain/config/tmp_genesis.json && mv $HOME/.seda-chain/config/tmp_genesis.json $HOME/.seda-chain/config/genesis.json
cat $HOME/.seda-chain/config/genesis.json | jq '.app_state["gov"]["params"]["voting_period"]="180s"' > $HOME/.seda-chain/config/tmp_genesis.json && mv $HOME/.seda-chain/config/tmp_genesis.json $HOME/.seda-chain/config/genesis.json
cat $HOME/.seda-chain/config/genesis.json | jq '.app_state["gov"]["params"]["expedited_voting_period"]="150s"' > $HOME/.seda-chain/config/tmp_genesis.json && mv $HOME/.seda-chain/config/tmp_genesis.json $HOME/.seda-chain/config/genesis.json
cat $HOME/.seda-chain/config/genesis.json | jq '.app_state["gov"]["params"]["max_deposit_period"]="180s"' > $HOME/.seda-chain/config/tmp_genesis.json && mv $HOME/.seda-chain/config/tmp_genesis.json $HOME/.seda-chain/config/genesis.json
cat $HOME/.seda-chain/config/genesis.json | jq '.app_state["gov"]["params"]["min_initial_deposit_ratio"]="0.010000000000000000"' > $HOME/.seda-chain/config/tmp_genesis.json && mv $HOME/.seda-chain/config/tmp_genesis.json $HOME/.seda-chain/config/genesis.json

# mint params
cat $HOME/.seda-chain/config/genesis.json | jq '.app_state["mint"]["params"]["blocks_per_year"]="4204800"' > $HOME/.seda-chain/config/tmp_genesis.json && mv $HOME/.seda-chain/config/tmp_genesis.json $HOME/.seda-chain/config/genesis.json

# slashing params
cat $HOME/.seda-chain/config/genesis.json | jq '.app_state["slashing"]["params"]["signed_blocks_window"]="10000"' > $HOME/.seda-chain/config/tmp_genesis.json && mv $HOME/.seda-chain/config/tmp_genesis.json $HOME/.seda-chain/config/genesis.json
cat $HOME/.seda-chain/config/genesis.json | jq '.app_state["slashing"]["params"]["min_signed_per_window"]="0.050000000000000000"' > $HOME/.seda-chain/config/tmp_genesis.json && mv $HOME/.seda-chain/config/tmp_genesis.json $HOME/.seda-chain/config/genesis.json
cat $HOME/.seda-chain/config/genesis.json | jq '.app_state["slashing"]["params"]["downtime_jail_duration"]="600s"' > $HOME/.seda-chain/config/tmp_genesis.json && mv $HOME/.seda-chain/config/tmp_genesis.json $HOME/.seda-chain/config/genesis.json
cat $HOME/.seda-chain/config/genesis.json | jq '.app_state["slashing"]["params"]["slash_fraction_double_sign"]="0.050000000000000000"' > $HOME/.seda-chain/config/tmp_genesis.json && mv $HOME/.seda-chain/config/tmp_genesis.json $HOME/.seda-chain/config/genesis.json
cat $HOME/.seda-chain/config/genesis.json | jq '.app_state["slashing"]["params"]["slash_fraction_downtime"]="0.0001"' > $HOME/.seda-chain/config/tmp_genesis.json && mv $HOME/.seda-chain/config/tmp_genesis.json $HOME/.seda-chain/config/genesis.json

# consensus params
cat $HOME/.seda-chain/config/genesis.json | jq '.consensus["params"]["block"]["max_gas"]="100000000"' > $HOME/.seda-chain/config/tmp_genesis.json && mv $HOME/.seda-chain/config/tmp_genesis.json $HOME/.seda-chain/config/genesis.json

# TO-DO gov (intentionally adjusted for testing): voting_params.voting_period, params.voting_period, params.expedited_voting_period, min_deposit[0].amount, max_deposit_period
# TO-DO wasm params

#
#   ADD GENESIS ACCOUNTS
#
for i in ${!GENESIS_ADDRESSES[@]}; do
    $LOCAL_BIN add-genesis-account ${GENESIS_ADDRESSES[$i]} 100000000000000000seda
done

set +u
if [ ! -z "$SATOSHI" ]; then
    $LOCAL_BIN add-genesis-account $SATOSHI 10000000000000000000seda
fi
if [ ! -z "$FAUCET" ]; then
    $LOCAL_BIN add-genesis-account $FAUCET 1000000000000000000seda
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

    # to create their gentx
    $LOCAL_BIN add-genesis-account $VALIDATOR_ADDRESS 500000000000000000seda --home $INDIVIDUAL_VAL_HOME_DIR
    # to output geneis file
    $LOCAL_BIN add-genesis-account $VALIDATOR_ADDRESS 500000000000000000seda

    $LOCAL_BIN gentx ${MONIKERS[$i]} ${SELF_DELEGATION_AMOUNTS[$i]} --moniker=${MONIKERS[$i]} --keyring-backend=test --home $INDIVIDUAL_VAL_HOME_DIR --ip=${IPS[$i]} --chain-id $CHAIN_ID

    cp -a $INDIVIDUAL_VAL_CONFIG_DIR/gentx/. $GENTX_DIR
done

cp -r $GENTX_DIR $HOME_CONFIG_DIR
$LOCAL_BIN collect-gentxs --home $HOME_DIR
cp $HOME_CONFIG_DIR/genesis.json $NODE_DIR
