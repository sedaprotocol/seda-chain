#!/bin/bash
set -e

#
# This script is run on a node to configure cosmovisor, shared library,
# and systemctl service.
#
# NOTE: Assumes ami-0a1ab4a3fcf997a9d
WASMVM_VERSION=$1

ARCH=$(uname -m)
if [ $ARCH != "aarch64" ]; then
	ARCH="x86_64"
fi

COSMOVISOR_URL=https://github.com/cosmos/cosmos-sdk/releases/download/cosmovisor%2Fv1.3.0/cosmovisor-v1.3.0-linux-amd64.tar.gz
if [ $ARCH = "aarch64" ]; then
	COSMOVISOR_URL=https://github.com/cosmos/cosmos-sdk/releases/download/cosmovisor%2Fv1.3.0/cosmovisor-v1.3.0-linux-arm64.tar.gz
fi
LIBWASMVM_URL=https://github.com/CosmWasm/wasmvm/releases/download/$WASMVM_VERSION/libwasmvm.$ARCH.so

COSMOS_LDS=$HOME/COSMOS_LDS
SYSFILE=/etc/systemd/system/seda-node.service

# set up cosmovisor if necessary
if ! which cosmovisor >/dev/null; then
	printf "\n\n\nSETTING UP COSMOVISOR\n\n\n\n"

	curl -LO $COSMOVISOR_URL
	mkdir -p tmp
	tar -xzvf $(basename $COSMOVISOR_URL) -C ./tmp
	sudo mv ./tmp/cosmovisor /usr/local/bin
	rm -rf ./tmp

	echo 'export DAEMON_NAME=sedad' >> $HOME/.bashrc
	echo 'export DAEMON_HOME=$HOME/.sedad' >> $HOME/.bashrc
	echo 'export DAEMON_DATA_BACKUP_DIR=$HOME/.sedad' >> $HOME/.bashrc
	echo 'export DAEMON_ALLOW_DOWNLOAD_BINARIES=false' >> $HOME/.bashrc
	echo 'export DAEMON_RESTART_AFTER_UPGRADE=true' >> $HOME/.bashrc
	echo 'export UNSAFE_SKIP_BACKUP=false' >> $HOME/.bashrc
	echo 'export DAEMON_POLL_INTERVAL=300ms' >> $HOME/.bashrc
	echo 'export DAEMON_RESTART_DELAY=30s' >> $HOME/.bashrc
	echo 'export DAEMON_LOG_BUFFER_SIZE=512' >> $HOME/.bashrc
	echo 'export DAEMON_PREUPGRADE_MAX_RETRIES=0' >> $HOME/.bashrc
	echo 'export PATH=$PATH:$HOME/.sedad/cosmovisor/current/bin' >> $HOME/.bashrc

	source $HOME/.bashrc
fi

# set up shared libraries (overwrite if already exists to ensure desired version)
printf "\n\n\nSETTING UP SHARED LIBRARY\n\n\n\n"
mkdir -p $COSMOS_LDS
curl -LO $LIBWASMVM_URL
mv $(basename $LIBWASMVM_URL) $COSMOS_LDS
echo 'export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:$HOME/COSMOS_LDS' >> $HOME/.bashrc

# create systemctl service file if necessary
if [ ! -f $SYSFILE ]; then
printf "\n\n\nSETTING UP SYSTEMCTL\n\n\n\n"

sudo tee /etc/systemd/system/seda-node.service > /dev/null <<EOF
[Unit]
Description=Seda Node Service
After=network-online.target

[Service]
Environment="DAEMON_NAME=sedad"
Environment="DAEMON_HOME=$HOME/.sedad"
Environment="DAEMON_DATA_BACKUP_DIR=$HOME/.sedad"

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

