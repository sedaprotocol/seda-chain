#!/bin/bash
set -e
set -x

# always returns true so set -e doesn't exit if it is not running.
killall sedad || true
rm -rf $HOME/.sedad/

# make four chain directories
mkdir $HOME/.sedad
mkdir $HOME/.sedad/validator1
mkdir $HOME/.sedad/validator2
# mkdir $HOME/.sedad/validator3
# mkdir $HOME/.sedad/validator4

# init all four validators
sedad init --default-denom aseda --chain-id=testing validator1 --home=$HOME/.sedad/validator1
sedad init --default-denom aseda --chain-id=testing validator2 --home=$HOME/.sedad/validator2
# sedad init --default-denom aseda --chain-id=testing validator3 --home=$HOME/.sedad/validator3
# sedad init --default-denom aseda --chain-id=testing validator4 --home=$HOME/.sedad/validator4
# create keys for all four validators
sedad keys add validator1 --keyring-backend=test --home=$HOME/.sedad/validator1
sedad keys add validator2 --keyring-backend=test --home=$HOME/.sedad/validator2
# sedad keys add validator3 --keyring-backend=test --home=$HOME/.sedad/validator3
# sedad keys add validator4 --keyring-backend=test --home=$HOME/.sedad/validator4



# create validator node with tokens to transfer to the four other nodes
sedad add-genesis-account $(sedad keys show validator1 -a --keyring-backend=test --home=$HOME/.sedad/validator1) 10000000000000000000000aseda --home=$HOME/.sedad/validator1
sedad gentx validator1 10000000000000000000aseda --keyring-backend=test --home=$HOME/.sedad/validator1 --chain-id=testing
sedad collect-gentxs --home=$HOME/.sedad/validator1


# port key (validator1 uses default ports)
# validator1 1317, 9050, 9091, 26658, 26657, 26656, 6060, 26660
# validator2 1316, 9088, 9089, 26655, 26654, 26653, 6061, 26630
# validator3 1315, 9086, 9087, 26652, 26651, 26650, 6062, 26620
# validator4 1314, 9084, 9085, 26649, 26648, 26647, 6063, 26610


# change app.toml values
VALIDATOR1_APP_TOML=$HOME/.sedad/validator1/config/app.toml
VALIDATOR2_APP_TOML=$HOME/.sedad/validator2/config/app.toml
# VALIDATOR3_APP_TOML=$HOME/.sedad/validator3/config/app.toml
# VALIDATOR4_APP_TOML=$HOME/.sedad/validator4/config/app.toml

# validator1
sed -i -E 's|0.0.0.0:9090|0.0.0.0:9050|g' $VALIDATOR1_APP_TOML

# validator2
sed -i -E 's|tcp://0.0.0.0:1317|tcp://0.0.0.0:1316|g' $VALIDATOR2_APP_TOML
sed -i -E 's|0.0.0.0:9090|0.0.0.0:9088|g' $VALIDATOR2_APP_TOML
sed -i -E 's|0.0.0.0:9091|0.0.0.0:9089|g' $VALIDATOR2_APP_TOML

# # validator3
# sed -i -E 's|tcp://0.0.0.0:1317|tcp://0.0.0.0:1315|g' $VALIDATOR3_APP_TOML
# sed -i -E 's|0.0.0.0:9090|0.0.0.0:9086|g' $VALIDATOR3_APP_TOML
# sed -i -E 's|0.0.0.0:9091|0.0.0.0:9087|g' $VALIDATOR3_APP_TOML
# sed -i -E 's|adaptive-fee-enabled = "false"|adaptive-fee-enabled = "true"|g' $VALIDATOR3_APP_TOML

# # validator4
# sed -i -E 's|tcp://0.0.0.0:1317|tcp://0.0.0.0:1314|g' $VALIDATOR4_APP_TOML
# sed -i -E 's|0.0.0.0:9090|0.0.0.0:9084|g' $VALIDATOR4_APP_TOML
# sed -i -E 's|0.0.0.0:9091|0.0.0.0:9085|g' $VALIDATOR4_APP_TOML
# sed -i -E 's|adaptive-fee-enabled = "false"|adaptive-fee-enabled = "true"|g' $VALIDATOR4_APP_TOML


# change config.toml values
VALIDATOR1_CONFIG=$HOME/.sedad/validator1/config/config.toml
VALIDATOR2_CONFIG=$HOME/.sedad/validator2/config/config.toml
# VALIDATOR3_CONFIG=$HOME/.sedad/validator3/config/config.toml
# VALIDATOR4_CONFIG=$HOME/.sedad/validator4/config/config.toml

# validator1
sed -i -E 's|allow_duplicate_ip = false|allow_duplicate_ip = true|g' $VALIDATOR1_CONFIG
# sed -i -E 's|version = "v0"|version = "v1"|g' $VALIDATOR1_CONFIG
sed -i -E 's|prometheus = false|prometheus = true|g' $VALIDATOR1_CONFIG

# validator2
sed -i -E 's|tcp://127.0.0.1:26658|tcp://127.0.0.1:26655|g' $VALIDATOR2_CONFIG
sed -i -E 's|tcp://0.0.0.0:26657|tcp://0.0.0.0:26654|g' $VALIDATOR2_CONFIG
sed -i -E 's|tcp://0.0.0.0:26656|tcp://0.0.0.0:26653|g' $VALIDATOR2_CONFIG
sed -i -E 's|allow_duplicate_ip = false|allow_duplicate_ip = true|g' $VALIDATOR2_CONFIG
sed -i -E 's|prometheus = false|prometheus = true|g' $VALIDATOR2_CONFIG
sed -i -E 's|prometheus_listen_addr = ":26660"|prometheus_listen_addr = ":26630"|g' $VALIDATOR2_CONFIG

# # validator3
# sed -i -E 's|tcp://127.0.0.1:26658|tcp://127.0.0.1:26652|g' $VALIDATOR3_CONFIG
# sed -i -E 's|tcp://0.0.0.0:26657|tcp://0.0.0.0:26651|g' $VALIDATOR3_CONFIG
# sed -i -E 's|tcp://0.0.0.0:26656|tcp://0.0.0.0:26650|g' $VALIDATOR3_CONFIG
# sed -i -E 's|allow_duplicate_ip = false|allow_duplicate_ip = true|g' $VALIDATOR3_CONFIG
# sed -i -E 's|prometheus = false|prometheus = true|g' $VALIDATOR3_CONFIG
# sed -i -E 's|prometheus_listen_addr = ":26660"|prometheus_listen_addr = ":26620"|g' $VALIDATOR3_CONFIG

# # validator4
# sed -i -E 's|tcp://127.0.0.1:26658|tcp://127.0.0.1:26649|g' $VALIDATOR4_CONFIG
# sed -i -E 's|tcp://0.0.0.0:26657|tcp://0.0.0.0:26648|g' $VALIDATOR4_CONFIG
# sed -i -E 's|tcp://0.0.0.0:26656|tcp://0.0.0.0:26647|g' $VALIDATOR4_CONFIG
# sed -i -E 's|allow_duplicate_ip = false|allow_duplicate_ip = true|g' $VALIDATOR4_CONFIG
# sed -i -E 's|prometheus = false|prometheus = true|g' $VALIDATOR4_CONFIG
# sed -i -E 's|prometheus_listen_addr = ":26660"|prometheus_listen_addr = ":26610"|g' $VALIDATOR4_CONFIG


# copy validator1 genesis file to validator2-4
cp $HOME/.sedad/validator1/config/genesis.json $HOME/.sedad/validator2/config/genesis.json
# cp $HOME/.sedad/validator1/config/genesis.json $HOME/.sedad/validator3/config/genesis.json
# cp $HOME/.sedad/validator1/config/genesis.json $HOME/.sedad/validator4/config/genesis.json


# copy tendermint node id of validator1 to persistent peers of validator2-4
sed -i -E "s|persistent_peers = \"\"|persistent_peers = \"$(sedad tendermint show-node-id --home=$HOME/.sedad/validator1)@localhost:26656\"|g" $HOME/.sedad/validator2/config/config.toml
# sed -i -E "s|persistent_peers = \"\"|persistent_peers = \"$(sedad tendermint show-node-id --home=$HOME/.sedad/validator1)@localhost:26656\"|g" $HOME/.sedad/validator3/config/config.toml
# sed -i -E "s|persistent_peers = \"\"|persistent_peers = \"$(sedad tendermint show-node-id --home=$HOME/.sedad/validator1)@localhost:26656\"|g" $HOME/.sedad/validator4/config/config.toml

# start all four validators
tmux new-session -s validator1 -d 
tmux send -t validator1 'sedad start --home=$HOME/.sedad/validator1 2>&1 | tee local_chain.log' ENTER

tmux new-session -s validator2 -d 
tmux send -t validator2 'sedad start --home=$HOME/.sedad/validator2 2>&1 | tee local_chain_2.log' ENTER

# tmux new-session -s validator3 -d 
# tmux send -t validator3 'sedad start --home=$HOME/.sedad/validator3' ENTER

# tmux new-session -s validator4 -d 
# tmux send -t validator4 'sedad start --home=$HOME/.sedad/validator4' ENTER

# send aseda from first validator to second validator
echo "Waiting 15 seconds to send funds to validators 2, 3, and 4..."
sleep 15
sedad tx bank send validator1 $(sedad keys show validator2 -a --keyring-backend=test --home=$HOME/.sedad/validator2) 50000000000000000000aseda --keyring-backend=test --home=$HOME/.sedad/validator1 --chain-id=testing --node http://localhost:26657 --yes --gas-prices 10000000000aseda --gas auto --gas-adjustment 2.0
# sedad tx bank send validator1 $(sedad keys show validator3 -a --keyring-backend=test --home=$HOME/.sedad/validator3) 400000000aseda --keyring-backend=test --home=$HOME/.sedad/validator1 --chain-id=testing --node http://localhost:26657 --yes --gas-prices 10000000000aseda --gas auto --gas-adjustment 2.0
# sedad tx bank send validator1 $(sedad keys show validator4 -a --keyring-backend=test --home=$HOME/.sedad/validator4) 400000000aseda --keyring-backend=test --home=$HOME/.sedad/validator1 --chain-id=testing --node http://localhost:26657 --yes --gas-prices 10000000000aseda --gas auto --gas-adjustment 2.0
sleep 15

# create second, third and fourth validator
# --pubkey=$(sedad tendermint show-validator --home=$HOME/.sedad/validator2)
VAL2_PUBKEY=$(sedad tendermint show-validator --home=$HOME/.sedad/validator2)
jq --argjson asdf "$VAL2_PUBKEY" '.pubkey = $asdf' validator2.json > validator2_out.json

sedad tx staking create-validator validator2.json --from=validator2 --chain-id="testing" --keyring-backend=test --home=$HOME/.sedad/validator2 --node http://localhost:26657 --yes --gas-prices 10000000000aseda --gas auto --gas-adjustment 2.0

# sedad tx staking create-validator --amount=500000000aseda --from=validator2 --pubkey=$(sedad tendermint show-validator --home=$HOME/.sedad/validator2) --moniker="validator2" --chain-id="testing" --commission-rate="0.1" --commission-max-rate="0.2" --commission-max-change-rate="0.05" --min-self-delegation="500000000" --keyring-backend=test --home=$HOME/.sedad/validator2 --broadcast-mode async --node http://localhost:26657 --yes --fees 1000000aseda
# sedad tx staking create-validator --amount=400000000aseda --from=validator3 --pubkey=$(sedad tendermint show-validator --home=$HOME/.sedad/validator3) --moniker="validator3" --chain-id="testing" --commission-rate="0.1" --commission-max-rate="0.2" --commission-max-change-rate="0.05" --min-self-delegation="400000000" --keyring-backend=test --home=$HOME/.sedad/validator3 --broadcast-mode async --node http://localhost:26657 --yes --fees 1000000aseda
# sedad tx staking create-validator --amount=400000000aseda --from=validator4 --pubkey=$(sedad tendermint show-validator --home=$HOME/.sedad/validator4) --moniker="validator4" --chain-id="testing" --commission-rate="0.1" --commission-max-rate="0.2" --commission-max-change-rate="0.05" --min-self-delegation="400000000" --keyring-backend=test --home=$HOME/.sedad/validator4 --broadcast-mode async --node http://localhost:26657 --yes --fees 1000000aseda

echo "All 4 Validators are up and running!"
