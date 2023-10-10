#!/bin/bash
set -euxo pipefail


# NOTE:
# Assuming systemctl is set up for seda-node.service
# Assuming cosmovisor has been set up
# 
# To FIX:
# connection closing after every ssh command
#

#
#   PARAMETERS
#
SSH_KEY=~/.ssh/id_rsa # key used for ssh
BIN=../../build/seda-chaind # chain binary executable on your machine
LINUX_BIN=../../build/seda-chaind-linux # linux version of chain binary
NODE_DIR=./nodes # OUT_DIR in other scripts
DESTINATIONS=(
	"ec2-user@ec2-18-169-59-167.eu-west-2.compute.amazonaws.com"
	"ec2-user@ec2-35-178-98-62.eu-west-2.compute.amazonaws.com"
)
IPS=(
	"18.169.59.167:26656"
	"35.178.98.62:26656"
)


# Prelimiary checks
if [ ! -f "$SSH_KEY" ]; then
  echo "ssh key file not found."
  exit 1
fi
if [ ! -f "$BIN" ]; then
  echo "binary file not found."
  exit 1
fi
if [ ! -f "$LINUX_BIN" ]; then
  echo "linux binary file not found."
  exit 1
fi

SEEDS=()
for i in ${!DESTINATIONS[@]}; do
	SEED=$($BIN tendermint show-node-id --home ./$NODE_DIR/node$i)
	SEEDS+=("$SEED@${IPS[$i]}")
done

printf -v list '%s,' "${SEEDS[@]}"
SEEDS_LIST="${list%,}"
echo $SEEDS_LIST

for i in ${!DESTINATIONS[@]}; do
	cp $NODE_DIR/genesis.json $NODE_DIR/node$i/config/genesis.json

	sed -i '' "s/seeds = \"\"/seeds = \"${SEEDS_LIST}\"/g" ./$NODE_DIR/node$i/config/config.toml

	# stop and remove
	ssh -i $SSH_KEY -t ${DESTINATIONS[$i]} 'sudo systemctl stop seda-node.service'
	ssh -i $SSH_KEY -t ${DESTINATIONS[$i]} 'sudo rm -f /var/log/seda-chain-error.log'
	ssh -i $SSH_KEY -t ${DESTINATIONS[$i]} 'sudo rm -f /var/log/seda-chain-output.log'

	ssh -i $SSH_KEY -t ${DESTINATIONS[$i]} 'sudo rm -rf /home/ec2-user/.seda-chain'

	# upload
	scp -i $SSH_KEY -r ./$NODE_DIR/node$i ${DESTINATIONS[$i]}:/home/ec2-user/.seda-chain

	ssh -i $SSH_KEY -t ${DESTINATIONS[$i]} 'mkdir -p /home/ec2-user/.seda-chain/cosmovisor/genesis/bin'
	scp -i $SSH_KEY $LINUX_BIN ${DESTINATIONS[$i]}:/home/ec2-user/.seda-chain/cosmovisor/genesis/bin/seda-chaind

	# start
	ssh -i $SSH_KEY -t ${DESTINATIONS[$i]} 'chmod 755 /home/ec2-user/.seda-chain/cosmovisor/genesis/bin/seda-chaind'
	ssh -i $SSH_KEY -t ${DESTINATIONS[$i]} 'sudo systemctl daemon-reload'
	ssh -i $SSH_KEY -t ${DESTINATIONS[$i]} 'sudo systemctl start seda-node.service'
done
