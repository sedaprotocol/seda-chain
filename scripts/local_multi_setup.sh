#!/bin/bash
set -ex

# The script requires tmux.
if ! command -v tmux &> /dev/null; then
    echo "Error: tmux is not installed. Please install tmux and try again."
    exit 1
fi

make build
BIN=./build/sedad

# always returns true so set -e doesn't exit if it is not running.
killall sedad || true
tmux kill-session -t validator1 || true
tmux kill-session -t validator2 || true
tmux kill-session -t validator3 || true
tmux kill-session -t validator4 || true
rm -rf $HOME/.sedad/
rm -f val1_multi_local.log
rm -f val2_multi_local.log
rm -f val3_multi_local.log
rm -f val4_multi_local.log

# make four chain directories
mkdir $HOME/.sedad
mkdir $HOME/.sedad/validator1
mkdir $HOME/.sedad/validator2
mkdir $HOME/.sedad/validator3
mkdir $HOME/.sedad/validator4

# init all four validators
$BIN init --default-denom aseda --chain-id=testing validator1 --home=$HOME/.sedad/validator1
$BIN init --default-denom aseda --chain-id=testing validator2 --home=$HOME/.sedad/validator2
$BIN init --default-denom aseda --chain-id=testing validator3 --home=$HOME/.sedad/validator3
$BIN init --default-denom aseda --chain-id=testing validator4 --home=$HOME/.sedad/validator4

# create keys for all four validators
$BIN keys add validator1 --keyring-backend=test --home=$HOME/.sedad/validator1
$BIN keys add validator2 --keyring-backend=test --home=$HOME/.sedad/validator2
$BIN keys add validator3 --keyring-backend=test --home=$HOME/.sedad/validator3
$BIN keys add validator4 --keyring-backend=test --home=$HOME/.sedad/validator4

# create validator node with tokens to transfer to the three other nodes
sed -i '' 's/allow-unencrypted-seda-keys = false/allow-unencrypted-seda-keys = true/' $HOME/.sedad/validator1/config/app.toml

$BIN add-genesis-account $($BIN keys show validator1 -a --keyring-backend=test --home=$HOME/.sedad/validator1) 100000000000000000000aseda --home=$HOME/.sedad/validator1
$BIN gentx validator1 10000000000000000000aseda --keyring-backend=test --home=$HOME/.sedad/validator1 --key-file-no-encryption --chain-id=testing 
$BIN collect-gentxs --home=$HOME/.sedad/validator1

# app.toml port key (validator1 uses default ports)
# validator  LCD,  gRPC
# validator1 1317, 9090
# validator2 1316, 9088
# validator3 1315, 9087
# validator4 1314, 9086

# change app.toml values
VALIDATOR2_APP_TOML=$HOME/.sedad/validator2/config/app.toml
VALIDATOR3_APP_TOML=$HOME/.sedad/validator3/config/app.toml
VALIDATOR4_APP_TOML=$HOME/.sedad/validator4/config/app.toml

# validator2
sed -i '' -E 's|tcp://0.0.0.0:1317|tcp://0.0.0.0:1316|g' $VALIDATOR2_APP_TOML # API server
sed -i '' -E 's|0.0.0.0:9090|0.0.0.0:9088|g' $VALIDATOR2_APP_TOML # gRPC server
sed -i '' 's/allow-unencrypted-seda-keys = false/allow-unencrypted-seda-keys = true/' $VALIDATOR2_APP_TOML
# validator3
sed -i '' -E 's|tcp://0.0.0.0:1317|tcp://0.0.0.0:1315|g' $VALIDATOR3_APP_TOML # API server
sed -i '' -E 's|0.0.0.0:9090|0.0.0.0:9087|g' $VALIDATOR3_APP_TOML # gRPC server
sed -i '' 's/allow-unencrypted-seda-keys = false/allow-unencrypted-seda-keys = true/' $VALIDATOR3_APP_TOML
# validator4
sed -i '' -E 's|tcp://0.0.0.0:1317|tcp://0.0.0.0:1314|g' $VALIDATOR4_APP_TOML # API server
sed -i '' -E 's|0.0.0.0:9090|0.0.0.0:9086|g' $VALIDATOR4_APP_TOML # gRPC server
sed -i '' 's/allow-unencrypted-seda-keys = false/allow-unencrypted-seda-keys = true/' $VALIDATOR4_APP_TOML

# config.toml port key (validator1 uses default ports)
# validator  ABCI , RPC  , P2P  , Prometheus
# validator1 26658, 26657, 26656, 26660
# validator2 26655, 26654, 26653, 26630
# validator3 26652, 26651, 26650, 26620
# validator4 26649, 26648, 26647, 26610

# change config.toml values
VALIDATOR1_CONFIG=$HOME/.sedad/validator1/config/config.toml
VALIDATOR2_CONFIG=$HOME/.sedad/validator2/config/config.toml
VALIDATOR3_CONFIG=$HOME/.sedad/validator3/config/config.toml
VALIDATOR4_CONFIG=$HOME/.sedad/validator4/config/config.toml

# Allow duplicate IP addresses and enable Prometheus
# validator1
sed -i '' -E 's|allow_duplicate_ip = false|allow_duplicate_ip = true|g' $VALIDATOR1_CONFIG
sed -i '' -E 's|prometheus = false|prometheus = true|g' $VALIDATOR1_CONFIG

# validator2
sed -i '' -E 's|allow_duplicate_ip = false|allow_duplicate_ip = true|g' $VALIDATOR2_CONFIG
sed -i '' -E 's|prometheus = false|prometheus = true|g' $VALIDATOR2_CONFIG
sed -i '' -E 's|prometheus_listen_addr = ":26660"|prometheus_listen_addr = ":26630"|g' $VALIDATOR2_CONFIG

# validator3
sed -i '' -E 's|allow_duplicate_ip = false|allow_duplicate_ip = true|g' $VALIDATOR3_CONFIG
sed -i '' -E 's|prometheus = false|prometheus = true|g' $VALIDATOR3_CONFIG
sed -i '' -E 's|prometheus_listen_addr = ":26660"|prometheus_listen_addr = ":26620"|g' $VALIDATOR3_CONFIG

# validator4
sed -i '' -E 's|allow_duplicate_ip = false|allow_duplicate_ip = true|g' $VALIDATOR4_CONFIG
sed -i '' -E 's|prometheus = false|prometheus = true|g' $VALIDATOR4_CONFIG
sed -i '' -E 's|prometheus_listen_addr = ":26660"|prometheus_listen_addr = ":26610"|g' $VALIDATOR4_CONFIG

# change config.toml ports
# validator2
sed -i '' -E 's|tcp://127.0.0.1:26658|tcp://127.0.0.1:26655|g' $VALIDATOR2_CONFIG # ABCI app
sed -i '' -E 's|tcp://0.0.0.0:26657|tcp://127.0.0.1:26654|g' $VALIDATOR2_CONFIG # RPC listen
sed -i '' -E 's|tcp://0.0.0.0:26656|tcp://0.0.0.0:26653|g' $VALIDATOR2_CONFIG # incoming connections

# validator3
sed -i '' -E 's|tcp://127.0.0.1:26658|tcp://127.0.0.1:26652|g' $VALIDATOR3_CONFIG # ABCI app
sed -i '' -E 's|tcp://0.0.0.0:26657|tcp://127.0.0.1:26651|g' $VALIDATOR3_CONFIG # RPC listen
sed -i '' -E 's|tcp://0.0.0.0:26656|tcp://0.0.0.0:26650|g' $VALIDATOR3_CONFIG # incoming connections

# validator4
sed -i '' -E 's|tcp://127.0.0.1:26658|tcp://127.0.0.1:26649|g' $VALIDATOR4_CONFIG # ABCI app
sed -i '' -E 's|tcp://0.0.0.0:26657|tcp://127.0.0.1:26648|g' $VALIDATOR4_CONFIG # RPC listen
sed -i '' -E 's|tcp://0.0.0.0:26656|tcp://0.0.0.0:26647|g' $VALIDATOR4_CONFIG # incoming connections

# modify genesis file
jq '.consensus.params.block.max_gas="100000000"' $HOME/.sedad/validator1/config/genesis.json > temp.json && mv temp.json $HOME/.sedad/validator1/config/genesis.json
jq '.consensus.params.abci.vote_extensions_enable_height = "10"' $HOME/.sedad/validator1/config/genesis.json > temp.json && mv temp.json $HOME/.sedad/validator1/config/genesis.json
jq '.app_state["pubkey"]["params"]["activation_block_delay"]="15"' $HOME/.sedad/validator1/config/genesis.json > temp.json && mv temp.json $HOME/.sedad/validator1/config/genesis.json
jq '.app_state["gov"]["voting_params"]["voting_period"]="30s"' $HOME/.sedad/validator1/config/genesis.json > temp.json && mv temp.json $HOME/.sedad/validator1/config/genesis.json
jq '.app_state["gov"]["params"]["voting_period"]="30s"' $HOME/.sedad/validator1/config/genesis.json > temp.json && mv temp.json $HOME/.sedad/validator1/config/genesis.json
jq '.app_state["gov"]["params"]["expedited_voting_period"]="15s"' $HOME/.sedad/validator1/config/genesis.json > temp.json && mv temp.json $HOME/.sedad/validator1/config/genesis.json

cp $HOME/.sedad/validator1/config/genesis.json $HOME/.sedad/validator2/config/genesis.json
cp $HOME/.sedad/validator1/config/genesis.json $HOME/.sedad/validator3/config/genesis.json
cp $HOME/.sedad/validator1/config/genesis.json $HOME/.sedad/validator4/config/genesis.json

# copy tendermint node id of validator1 to persistent peers of validator2-4
NODE1_ID=$($BIN tendermint show-node-id --home=$HOME/.sedad/validator1 | tail -1)
sed -i '' -E "s|persistent_peers = \"\"|persistent_peers = \"${NODE1_ID}@localhost:26656\"|g" $HOME/.sedad/validator2/config/config.toml
sed -i '' -E "s|persistent_peers = \"\"|persistent_peers = \"${NODE1_ID}@localhost:26656\"|g" $HOME/.sedad/validator3/config/config.toml
sed -i '' -E "s|persistent_peers = \"\"|persistent_peers = \"${NODE1_ID}@localhost:26656\"|g" $HOME/.sedad/validator4/config/config.toml

# start all four validators
tmux new-session -s validator1 -d 
tmux send -t validator1 'BIN=./build/sedad' ENTER
tmux send -t validator1 '$BIN start --home=$HOME/.sedad/validator1 > val1_multi_local.log 2>&1 &' ENTER

tmux new-session -s validator2 -d 
tmux send -t validator2 'BIN=./build/sedad' ENTER

tmux new-session -s validator3 -d 
tmux send -t validator3 'BIN=./build/sedad' ENTER

tmux new-session -s validator4 -d 
tmux send -t validator4 'BIN=./build/sedad' ENTER

echo "begin sending funds to validators 2, 3, & 4"
sleep 10
$BIN tx bank send validator1 $($BIN keys show validator2 -a --keyring-backend=test --home=$HOME/.sedad/validator2) 20000000000000000000aseda --keyring-backend=test --home=$HOME/.sedad/validator1 --chain-id=testing --node http://localhost:26657 --broadcast-mode sync --yes --gas-prices 10000000000aseda --gas auto --gas-adjustment 1.7
sleep 10
$BIN tx bank send validator1 $($BIN keys show validator3 -a --keyring-backend=test --home=$HOME/.sedad/validator3) 20000000000000000000aseda --keyring-backend=test --home=$HOME/.sedad/validator1 --chain-id=testing --node http://localhost:26657 --broadcast-mode sync --yes --gas-prices 10000000000aseda --gas auto --gas-adjustment 1.7
sleep 10
$BIN tx bank send validator1 $($BIN keys show validator4 -a --keyring-backend=test --home=$HOME/.sedad/validator4) 20000000000000000000aseda --keyring-backend=test --home=$HOME/.sedad/validator1 --chain-id=testing --node http://localhost:26657 --broadcast-mode sync --yes --gas-prices 10000000000aseda --gas auto --gas-adjustment 1.7
sleep 10

echo "begin sending create validator txs"

cat << EOF > validator2.json
{
	"pubkey": $($BIN tendermint show-validator --home=$HOME/.sedad/validator2),
	"amount": "10000000000000000000aseda",
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
$BIN tx staking create-validator validator2.json --from=validator2 --keyring-backend=test --home=$HOME/.sedad/validator2 --broadcast-mode sync --chain-id=testing --node http://localhost:26657 --yes --gas-prices 10000000000aseda --gas auto --gas-adjustment 1.7 --key-file-no-encryption
rm validator2.json
tmux send -t validator2 '$BIN start --home=$HOME/.sedad/validator2 --log_level debug > val2_multi_local.log 2>&1 &' ENTER

cat << EOF > validator3.json
{
	"pubkey": $($BIN tendermint show-validator --home=$HOME/.sedad/validator3),
	"amount": "10000000000000000000aseda",
	"moniker": "validator3",
	"identity": "val3",
	"website": "val3.com",
	"security": "val3@yandex.kr",
	"details": "val3 details",
	"commission-rate": "0.1",
	"commission-max-rate": "0.2",
	"commission-max-change-rate": "0.01",
	"min-self-delegation": "1"
}
EOF
$BIN tx staking create-validator validator3.json --from=validator3 --keyring-backend=test --home=$HOME/.sedad/validator3 --broadcast-mode sync --chain-id=testing --node http://localhost:26657 --yes --gas-prices 10000000000aseda --gas auto --gas-adjustment 1.7 --key-file-no-encryption
rm validator3.json
tmux send -t validator3 '$BIN start --home=$HOME/.sedad/validator3 --log_level debug > val3_multi_local.log 2>&1 &' ENTER

cat << EOF > validator4.json
{
	"pubkey": $($BIN tendermint show-validator --home=$HOME/.sedad/validator4),
	"amount": "5000000000000000000aseda",
	"moniker": "validator4",
	"identity": "val4",
	"website": "val4.com",
	"security": "val4@yandex.kr",
	"details": "val4 details",
	"commission-rate": "0.1",
	"commission-max-rate": "0.2",
	"commission-max-change-rate": "0.01",
	"min-self-delegation": "1"
}
EOF
$BIN tx staking create-validator validator4.json --from=validator4 --keyring-backend=test --home=$HOME/.sedad/validator4 --broadcast-mode sync --chain-id=testing --node http://localhost:26657 --yes --gas-prices 10000000000aseda --gas auto --gas-adjustment 1.7 --key-file-no-encryption
rm validator4.json
tmux send -t validator4 '$BIN start --home=$HOME/.sedad/validator4 --log_level debug > val4_multi_local.log 2>&1 &' ENTER

sleep 10
echo "4 validators are up and running!"
