#!/usr/bin/env sh
set -e

SEDA_DIR=$HOME/.sedad
CONFIG_DIR=$HOME/.sedad/config
DATA_DIR=$HOME/.sedad/data
NETWORK=${NETWORK:-testnet}

if
	[ ! -d "$SEDA_DIR/cosmovisor" ]
then
	echo "Creating cosmovisor directory and initializing"
	mkdir -p "$SEDA_DIR/cosmovisor"
	cosmovisor init /usr/local/bin/sedad
fi

if
	[ ! -f "$CONFIG_DIR/app.toml" ] || 
	[ ! -f "$CONFIG_DIR/client.toml" ] ||
	[ ! -f "$CONFIG_DIR/config.toml" ] ||
	[ ! -f "$CONFIG_DIR/genesis.json" ] ||
	[ ! -f "$CONFIG_DIR/node_key.json" ] ||
	[ ! -f "$CONFIG_DIR/priv_validator_key.json" ] ||
	[ ! -f "$DATA_DIR/priv_validator_state.json" ]
then
	if [ -z "$MONIKER" ]; then
		echo "MONIKER env not set"
		exit
	fi

	if [ -z "$NETWORK" ]; then
		echo "NETWORK env not set"
		exit
	fi

	echo "Initializing chain with moniker '$MONIKER' on chain $NETWORK..."
	sedad join "$MONIKER" --network $NETWORK
	echo "Initialization complete."
fi

exec "$@"