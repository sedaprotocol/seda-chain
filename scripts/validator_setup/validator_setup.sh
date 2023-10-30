#!/bin/bash
set -e

# Basic Setup Configurations
# --------------------------
# This script is used to setup a validator node for the seda-chain network
# It takes the following parameters insde a env file:
# 1. MNEMONIC: The mnemonic of the validator account
# 2. KEYRING_PASSWORD: The password to encrypt the keyring
# 3. NETWORK_ID: The network id of the seda-chain network
# 4. MONIKER: The moniker of the validator node
# 5. NODE_ADDRESS: The address of the seda-chain network
# IE:
# MONIKER="somemonicer"
 #MNEMONIC="a mnemonic"
 #KEYRING_PASSWORD="somepassword"
 #NETWORK_ID="devnet"
 #NODE_ADDRESS="http://35.177.180.184:26657"

BIN=seda-chaind # chain binary executable on your machine
KEY_NAME="${KEY_NAME:-default_key}"

#if NODE_ADDRESS is not provided exit with error
if [[ -z "${NODE_ADDRESS}" ]]; then
  echo "Error no key NODE_ADDRESS provided"
  exit 1
fi

if [[ -z "${KEYRING_PASSWORD}" ]]; then
  echo "Error no key  KEYRING_PASSWORD provided"
  exit 1
fi

if [[  -z "${NETWORK_ID}" ]]; then
  echo "Error no key NETWORK_ID provided"
  exit 1
fi

if [[  -z "${MONIKER}" ]]; then
  echo "Error no key password provided"
  exit 1
fi

if [[  -z "${MNEMONIC}" ]]; then
  echo "Error no key MNEMONIC provided"
  exit 1
fi

# Set the keyring-backend to file
# Initialize NODE config
echo "Initializing Node ..."

# Check if configuration directory seda-chain config directory exist if it does not
# exist initialize the node with the given MNEMONIC, MONIKER and NETWORK_ID
if ! [ -f /seda-chain/.seda-chain/config/genesis.json ]; then
    echo "Setting Up seda configuration"
    $BIN config keyring-backend file                             # use file backend
    echo $MNEMONIC | $BIN init join ${MONIKER} --network ${NETWORK_ID} --recover
  else
    echo "seda configuration already exists"
fi

echo "Node Initialized !"


# It creates a Client to the seda-chain network
echo "Connecting to Network ..."
echo $NODE_ADDRESS | $BIN config node $NODE_ADDRESS
echo "Connected to network !"

# Run node
echo "Running Node ..."
$BIN start