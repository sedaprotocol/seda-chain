#!/bin/bash

make install
make build
BIN=./build/seda-chaind

$BIN tendermint unsafe-reset-all
rm -rf ~/.seda-chain
$BIN init new node0

$BIN keys add satoshi --keyring-backend test
$BIN add-genesis-account $($BIN keys show satoshi --keyring-backend test -a) 10000000000000000seda

$BIN keys add acc1 --keyring-backend test
$BIN add-genesis-account $($BIN keys show acc1 --keyring-backend test -a) 10000000000000000seda


$BIN gentx satoshi 10000000000000000seda --keyring-backend test
$BIN collect-gentxs
$BIN start