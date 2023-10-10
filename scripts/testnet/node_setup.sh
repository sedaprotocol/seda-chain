#!/bin/bash
set -e

#
# This script set up a node for chain deployment by configuring
# cosmovisor, shared library, and systemctl service.
#
# NOTE: Assumes ami-0a1ab4a3fcf997a9d

COSMOVISOR_URL=https://github.com/cosmos/cosmos-sdk/releases/download/cosmovisor%2Fv1.3.0/cosmovisor-v1.3.0-linux-arm64.tar.gz
COSMOVISOR_TAR_GZ=cosmovisor-v1.3.0-linux-arm64.tar.gz
LIBWASMVM_URL=https://github.com/CosmWasm/wasmvm/releases/download/v1.3.0/libwasmvm.aarch64.so
LIBWASMVM=libwasmvm.aarch64.so

COSMOS_LDS=$HOME/COSMOS_LDS
SYSFILE=/etc/systemd/system/seda-node.service

# set up cosmovisor if necessary
if ! which cosmovisor >/dev/null; then
	printf "\n\n\nSETTING UP COSMOVISOR\n\n\n\n"

	curl -LO $COSMOVISOR_URL
	mkdir -p tmp
	tar -xzvf $COSMOVISOR_TAR_GZ -C ./tmp
	mv ./tmp/cosmovisor .
	rm -rf ./tmp ./$COSMOVISOR_TAR_GZ

	sudo mv $HOME/cosmovisor /usr/local/bin

	echo 'export DAEMON_NAME=seda-chaind' >> $HOME/.bashrc
	echo 'export DAEMON_HOME=$HOME/.seda-chain' >> $HOME/.bashrc
	echo 'export DAEMON_DATA_BACKUP_DIR=$HOME/.seda-chain' >> $HOME/.bashrc
	echo 'export DAEMON_ALLOW_DOWNLOAD_BINARIES=false' >> $HOME/.bashrc
	echo 'export DAEMON_RESTART_AFTER_UPGRADE=true' >> $HOME/.bashrc
	echo 'export UNSAFE_SKIP_BACKUP=false' >> $HOME/.bashrc
	echo 'export DAEMON_POLL_INTERVAL=300ms' >> $HOME/.bashrc
	echo 'export DAEMON_RESTART_DELAY=30s' >> $HOME/.bashrc
	echo 'export DAEMON_LOG_BUFFER_SIZE=512' >> $HOME/.bashrc
	echo 'export DAEMON_PREUPGRADE_MAX_RETRIES=0' >> $HOME/.bashrc
	echo 'export PATH=$PATH:$HOME/.seda-chain/cosmovisor/current/bin' >> $HOME/.bashrc

	# set up shared libraries if necessary
	if [ ! -d $COSMOS_LDS ]; then
		printf "\n\n\nSETTING UP SHARED LIBRARY\n\n\n\n"

		mkdir -p $COSMOS_LDS
		curl -LO $LIBWASMVM_URL
		mv $LIBWASMVM $COSMOS_LDS
		echo 'export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:$HOME/COSMOS_LDS' >> $HOME/.bashrc
	fi

	source $HOME/.bashrc
fi

# create systemctl service file if necessary
if [ ! -f $SYSFILE ]; then
printf "\n\n\nSETTING UP SYSTEMCTL\n\n\n\n"

sudo tee /etc/systemd/system/seda-node.service > /dev/null <<EOF
[Unit]
Description=Seda Node Service
After=network-online.target

[Service]
Environment="DAEMON_NAME=seda-chaind"
Environment="DAEMON_HOME=$HOME/.seda-chain"
Environment="DAEMON_DATA_BACKUP_DIR=$HOME/.seda-chain"

Environment="DAEMON_ALLOW_DOWNLOAD_BINARIES=false"
Environment="DAEMON_RESTART_AFTER_UPGRADE=true"
Environment="UNSAFE_SKIP_BACKUP=false"

Environment="DAEMON_POLL_INTERVAL=300ms"
Environment="DAEMON_RESTART_DELAY=30s"
Environment="DAEMON_LOG_BUFFER_SIZE=512"
Environment="DAEMON_PREUPGRADE_MAX_RETRIES=0"

Environment=LD_LIBRARY_PATH=/home/ec2-user/COSMOS_LDS

User=$USER
ExecStart=$(which cosmovisor) run start
Restart=always
RestartSec=3
LimitNOFILE=65535
LimitMEMLOCK=200M

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl enable seda-node
sudo systemctl daemon-reload
fi

echo done

