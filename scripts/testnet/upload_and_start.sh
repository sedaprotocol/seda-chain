#!/bin/bash
set -euxo pipefail


# NOTE:
# Assuming systemctl is set up for seda-node.service
# Assuming cosmovisor has been set up
# 
# To FIX:
# connection closing after every ssh command
#
source config.sh


################################################
############### Parameter checks ###############
################################################

# prelimiary checks on parameters
if [ $($LOCAL_BIN version) != $CHAIN_VERSION ]; then
    echo "Local chain version is" $($LOCAL_BIN version) "instead of" $CHAIN_VERSION
    exit 1
fi

if [ ! -f "$SSH_KEY" ]; then
  echo "ssh key file not found."
  exit 1
fi
if [ ! -f "$LOCAL_BIN" ]; then
  echo "local chain binary not found."
  exit 1
fi

# download chain binaries
curl -LO https://github.com/sedaprotocol/seda-chain/releases/download/$CHAIN_VERSION/seda-chaind-amd64
mv seda-chaind-amd64 $NODE_DIR
curl -LO https://github.com/sedaprotocol/seda-chain/releases/download/$CHAIN_VERSION/seda-chaind-arm64
mv seda-chaind-arm64 $NODE_DIR


################################################
############# Set up for new nodes #############
################################################

# upload setup script and run it
for i in ${!IPS[@]}; do
	scp -i $SSH_KEY -o StrictHostKeyChecking=no -r ./setup_node.sh ec2-user@${IPS[$i]}:/home/ec2-user
	ssh -i $SSH_KEY -t ec2-user@${IPS[$i]} "/home/ec2-user/setup_node.sh $WASMVM_VERSION"
	ssh -i $SSH_KEY -t ec2-user@${IPS[$i]} 'rm /home/ec2-user/setup_node.sh'
done


#################################################
############ Upload and start chain #############
#################################################

SEEDS=()
for i in ${!IPS[@]}; do
	SEED=$($LOCAL_BIN tendermint show-node-id --home $NODE_DIR/node$i)
	SEEDS+=("$SEED@${IPS[$i]}:26656")
done

printf -v list '%s,' "${SEEDS[@]}"
SEEDS_LIST="${list%,}"
echo $SEEDS_LIST

for i in ${!IPS[@]}; do
	cp $NODE_DIR/genesis.json $NODE_DIR/node$i/config/genesis.json

	if [[ "$OSTYPE" == "darwin"* ]]; then
		sed -i '' "s/seeds = \"\"/seeds = \"${SEEDS_LIST}\"/g" $NODE_DIR/node$i/config/config.toml
	else
		sed "s/seeds = \"\"/seeds = \"${SEEDS_LIST}\"/g" $NODE_DIR/node$i/config/config.toml > tmp
		cat tmp > $NODE_DIR/node$i/config/config.toml
		rm tmp
	fi

	# stop and remove
	ssh -i $SSH_KEY -t ec2-user@${IPS[$i]} 'sudo systemctl stop seda-node.service'
	ssh -i $SSH_KEY -t ec2-user@${IPS[$i]} 'sudo rm -f /var/log/seda-chain-error.log'
	ssh -i $SSH_KEY -t ec2-user@${IPS[$i]} 'sudo rm -f /var/log/seda-chain-output.log'

	ssh -i $SSH_KEY -t ec2-user@${IPS[$i]} 'sudo rm -rf /home/ec2-user/.seda-chain'

	# upload node files
	scp -i $SSH_KEY -r $NODE_DIR/node$i ec2-user@${IPS[$i]}:/home/ec2-user/.seda-chain

	# upload chain binary built for the corresponding architecture
	LINUX_BIN=$NODE_DIR/seda-chaind-amd64
	ARCH=$(ssh -i $SSH_KEY -t ec2-user@${IPS[$i]} 'uname -m') # aarch64 or x86_64
	if [ $ARCH == "aarch64" ]; then
		LINUX_BIN=$NODE_DIR/seda-chaind-arm64
	fi
	
	ssh -i $SSH_KEY -t ec2-user@${IPS[$i]} 'mkdir -p /home/ec2-user/.seda-chain/cosmovisor/genesis/bin'
	scp -i $SSH_KEY $LINUX_BIN ec2-user@${IPS[$i]}:/home/ec2-user/.seda-chain/cosmovisor/genesis/bin/seda-chaind

	# start
	ssh -i $SSH_KEY -t ec2-user@${IPS[$i]} 'chmod 755 /home/ec2-user/.seda-chain/cosmovisor/genesis/bin/seda-chaind'
	ssh -i $SSH_KEY -t ec2-user@${IPS[$i]} 'sudo systemctl daemon-reload'
	ssh -i $SSH_KEY -t ec2-user@${IPS[$i]} 'sudo systemctl start seda-node.service'
done

echo "Script has finished running."
