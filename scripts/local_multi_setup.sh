#!/bin/bash
set -e
set -x

make build
BIN=./build/sedad

# always returns true so set -e doesn't exit if it is not running.
killall sedad || true
rm -rf $HOME/.sedad/

# make four chain directories
mkdir $HOME/.sedad
mkdir $HOME/.sedad/validator1
mkdir $HOME/.sedad/validator2

# init all four validators
$BIN init --default-denom aseda --chain-id=testing validator1 --home=$HOME/.sedad/validator1
$BIN init --default-denom aseda --chain-id=testing validator2 --home=$HOME/.sedad/validator2

# create keys for all four validators
$BIN keys add validator1 --keyring-backend=test --home=$HOME/.sedad/validator1
$BIN keys add validator2 --keyring-backend=test --home=$HOME/.sedad/validator2

# create validator node with tokens to transfer to the four other nodes
$BIN add-genesis-account $($BIN keys show validator1 -a --keyring-backend=test --home=$HOME/.sedad/validator1) 10000000000000000000aseda --home=$HOME/.sedad/validator1
$BIN gentx validator1 1000000000000000000aseda --keyring-backend=test --home=$HOME/.sedad/validator1 --chain-id=testing
$BIN collect-gentxs --home=$HOME/.sedad/validator1

# port key (validator1 uses default ports)
# validator1 1317, 9050, 9091, 26658, 26657, 26656, 6060, 26660
# validator2 1316, 9088, 9089, 26655, 26654, 26653, 6061, 26630

# change app.toml values
VALIDATOR1_APP_TOML=$HOME/.sedad/validator1/config/app.toml
VALIDATOR2_APP_TOML=$HOME/.sedad/validator2/config/app.toml

# validator1
sed -i '' -E 's|0.0.0.0:9090|0.0.0.0:9050|g' $VALIDATOR1_APP_TOML

# validator2
sed -i '' -E 's|tcp://0.0.0.0:1317|tcp://0.0.0.0:1316|g' $VALIDATOR2_APP_TOML
sed -i '' -E 's|0.0.0.0:9090|0.0.0.0:9088|g' $VALIDATOR2_APP_TOML
sed -i '' -E 's|0.0.0.0:9091|0.0.0.0:9089|g' $VALIDATOR2_APP_TOML

# change config.toml values
VALIDATOR1_CONFIG=$HOME/.sedad/validator1/config/config.toml
VALIDATOR2_CONFIG=$HOME/.sedad/validator2/config/config.toml

# validator1
sed -i '' -E 's|allow_duplicate_ip = false|allow_duplicate_ip = true|g' $VALIDATOR1_CONFIG
# sed -i '' -E 's|version = "v0"|version = "v1"|g' $VALIDATOR1_CONFIG
sed -i '' -E 's|prometheus = false|prometheus = true|g' $VALIDATOR1_CONFIG

# validator2
sed -i '' -E 's|tcp://127.0.0.1:26658|tcp://127.0.0.1:26655|g' $VALIDATOR2_CONFIG
# sed -i '' -E 's|tcp://127.0.0.1:26657|tcp://127.0.0.1:26654|g' $VALIDATOR2_CONFIG
sed -i '' -E 's|tcp://0.0.0.0:26657|tcp://127.0.0.1:26654|g' $VALIDATOR2_CONFIG
sed -i '' -E 's|tcp://0.0.0.0:26656|tcp://0.0.0.0:26653|g' $VALIDATOR2_CONFIG
sed -i '' -E 's|allow_duplicate_ip = false|allow_duplicate_ip = true|g' $VALIDATOR2_CONFIG
sed -i '' -E 's|prometheus = false|prometheus = true|g' $VALIDATOR2_CONFIG
sed -i '' -E 's|prometheus_listen_addr = ":26660"|prometheus_listen_addr = ":26630"|g' $VALIDATOR2_CONFIG

# modify validator1 genesis file to validator2
jq '.consensus.params.abci.vote_extensions_enable_height = "15"' $HOME/.sedad/validator1/config/genesis.json > temp.json && mv temp.json $HOME/.sedad/validator1/config/genesis.json

cp $HOME/.sedad/validator1/config/genesis.json $HOME/.sedad/validator2/config/genesis.json

# copy tendermint node id of validator1 to persistent peers of validator2-4
sed -i '' -E "s|persistent_peers = \"\"|persistent_peers = \"$($BIN tendermint show-node-id --home=$HOME/.sedad/validator1)@localhost:26656\"|g" $HOME/.sedad/validator2/config/config.toml

# start all four validators
tmux new-session -s validator1 -d 
tmux send -t validator1 'BIN=./build/sedad' ENTER
tmux send -t validator1 '$BIN start --home=$HOME/.sedad/validator1' ENTER

tmux new-session -s validator2 -d 
tmux send -t validator2 'BIN=./build/sedad' ENTER
tmux send -t validator2 '$BIN start --home=$HOME/.sedad/validator2' ENTER

echo "Waiting 10 seconds to send funds to validators 2"
sleep 10
$BIN tx bank send validator1 $($BIN keys show validator2 -a --keyring-backend=test --home=$HOME/.sedad/validator2) 5000000000000000000aseda --keyring-backend=test --home=$HOME/.sedad/validator1 --chain-id=testing --node http://localhost:26657 --broadcast-mode sync --yes --gas-prices 10000000000aseda --gas auto --gas-adjustment 1.7

echo "Waiting 10 seconds to create validator 2"
sleep 10
cat << EOF > validator2.json
{
	"pubkey": $(./build/sedad tendermint show-validator --home=$HOME/.sedad/validator2),
	"amount": "1000000000000000000aseda",
	"moniker": "validator2",
	"identity": "val2",
	"website": "val2.com",
	"security": "val2@yandex.kr",
	"details": "val2 details",
	"commission-rate": "0.1",
	"commission-max-rate": "0.2",
	"commission-max-change-rate": "0.01",
	"min-self-delegation": "1"
}
EOF
$BIN tx staking create-validator validator2.json --from=validator2 --keyring-backend=test --home=$HOME/.sedad/validator2 --broadcast-mode sync --chain-id=testing --node http://localhost:26657 --yes --gas-prices 10000000000aseda --gas auto --gas-adjustment 1.7
rm validator2.json

echo "2 validators are up and running!"
sleep 10

# generate and register SEDA keys
$BIN tx pubkey add-seda-keys --from validator1 --keyring-backend=test --home=$HOME/.sedad/validator1 --gas-prices 10000000000aseda --gas auto --gas-adjustment 2.0 --keyring-backend test --chain-id=testing --node http://localhost:26657 --yes
$BIN tx pubkey add-seda-keys --from validator2 --keyring-backend=test --home=$HOME/.sedad/validator2 --gas-prices 10000000000aseda --gas auto --gas-adjustment 2.0 --keyring-backend test --chain-id=testing --node http://localhost:26657 --yes

# restart
sleep 10
tmux send -t validator1 'C-c'
tmux send -t validator1 '$BIN start --home=$HOME/.sedad/validator1' ENTER

tmux send -t validator2 'C-c'
tmux send -t validator2 '$BIN start --home=$HOME/.sedad/validator2' ENTER
