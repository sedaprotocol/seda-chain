#!/usr/bin/env bash
# Tell script to exit immediately if any cmd fails
set -e

# Basic Setup Configuration
# --------------------------
# This script is used to setup a node for the seda-chain network.


BIN=sedad
KEY_NAME="${KEY_NAME:-default_key}"

# Check if a env variable exists
# returns true if it doesn't exist false otherwise
function check_env_var {
  local var_name="$1"
  if [[ -z "${!var_name}" ]]; then
    echo "Error: no key $var_name provided"
    return 1
  fi
  return 0
}

# Define an error boolean
# Then check if required env variables are set
error=0
check_env_var "MONIKER" || error=1
check_env_var "NODE_ADDRESS" || error=1
check_env_var "NETWORK" || error=1

# TODO
# check_env_var "VALIDATOR" || error=1
# if [[ "$VALIDATOR" == "true" ]]; then
  # echo "Error: MNEMONIC is required for validator nodes since it requires funds"
  # check_env_var "MNEMONIC" || error=1
# fi

# If any one of them was missing then exit
if (( error )); then
  exit 1
fi

# Set the keyring backend to file
$BIN config keyring-backend file

function create_or_import_key {
  local key_name="$1"
  local mnemonic="$2"
  local recover_flag="$3"

  output=$(expect -c "
  set key_name \"$key_name\"
  set recover_flag \"$recover_flag\"
  set mnemonic \"$mnemonic\"
  spawn sedad keys add \$key_name \$recover_flag
  expect {
    \"> Enter your bip39 mnemonic\" {
      send \"$mnemonic\r\"
    }
    timeout {
      send_user \"Timed out waiting for enter mnemonic prompt\r\"
      exit 1
    }
  }
  expect {
    \"Enter keyring passphrase (attempt 1/3):\" {
      send \"$KEYRING_PASSWORD\r\"
      expect {
        \"Re-enter keyring passphrase:\" {
          send \"$KEYRING_PASSWORD\r\"
        }
        timeout {
          # We're done if we timeout after sending the passphrase
        }
      }
    }
    timeout {
      send_user \"Timed out waiting for enter passphrase prompt\r\"
      exit 1
    }
  }
  expect eof
")


  echo "$output"
}

# If the MNEMONIC is provided we import it
if [[ ! -z "${MNEMONIC}" ]]; then
  echo "Importing provided MNEMONIC..."
  create_or_import_key "$KEY_NAME" "$MNEMONIC" "--recover"
# If the MNEMONIC is not provided we generate one
else
  echo "Error no key MNEMONIC provided generating one..."
  echo "NOTE: This is done in the file backend..."
  echo "We recommend storing this somewhere secure..."
  output=$(create_or_import_key "$KEY_NAME" "" "")
  # We grab the mnemonic from the output.
  mnemonic=$(echo "$output" | awk '/Important/,0' | tail -n 1)
  # We check if the `env` file had a MNEMONIC variable in general...
  awk -v mnemonic="$mnemonic" 'BEGIN{OFS=FS="="} $1=="MNEMONIC"{$2=mnemonic}1' .env > .env.tmp && cat .env.tmp > .env && rm .env.tmp
  # Lastly set the env variable for this session.
  export MNEMONIC="$mnemonic"
fi


# Initialize NODE config
echo "Initializing Node ..."

# # Give docker image permission to write to the seda-chain config directory
# chmod -R a+w /seda-chain/.seda

# Check if configuration directory seda-chain config directory exist if it does not
# exist initialize the node with the given MNEMONIC, MONIKER and NETWORK
if ! [ -f /seda-chain/.seda/config/genesis.json ]; then
    echo "Setting Up seda configuration"
    echo $MNEMONIC | $BIN join ${MONIKER} --network ${NETWORK} --recover
  else
    echo "seda configuration already exists"
fi

echo "Node Initialized !"


# It creates a Client to the seda-chain network
echo "Setting client Network to `$NETWORK`..."
$BIN config node $NODE_ADDRESS

# Run node
echo "Running Node ..."
$BIN start
