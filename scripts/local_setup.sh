#!/bin/bash
set -e
set -x

#
# Local Single-node Setup
#
# NOTE: Run this script from project root.
#
make build
BIN=./build/seda-chaind

$BIN tendermint unsafe-reset-all
rm -rf ~/.seda-chain
$BIN init new node0

$BIN keys add satoshi --keyring-backend test
ADDR=$($BIN keys show satoshi --keyring-backend test -a)
$BIN add-genesis-account $ADDR 10000000000000000seda
$BIN gentx satoshi 10000000000000000seda --keyring-backend test
$BIN collect-gentxs
$BIN start
