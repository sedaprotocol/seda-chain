
NODE_DIR=./nodes # where node directories will be created
WASM_DIR=./artifacts # where Wasm files are located

HOME_DIR=$HOME/.seda-chain # chain directory
HOME_CONFIG_DIR=$HOME_DIR/config # chain config directory

BIN=$(git rev-parse --show-toplevel)/build/seda-chaind # chain binary executable on your machine
LINUX_BIN=$(git rev-parse --show-toplevel)/build/seda-chaind-linux # linux version of chain binary

# CHAIN_ID=seda-testnet
# GENESIS_TIME=


# Validators
# NOTE: Assumes 26656 port for p2p communication
# NOTE: Assumes user is ec2-user
# NOTE: The setup node script assumes ami-0a1ab4a3fcf997a9d
IPS=(
    "xx.xx.xx.xx"
    "xx.xx.xx.xx"
)
MONIKERS=(
    "node0"
    "node1"
)
SELF_DELEGATION_AMOUNTS=(
    "30000000000000000seda"
    "10000000000000000seda"
)

SSH_KEY=~/.ssh/id_rsa # key used for ssh


# Genesis acoounts addresses
# NOTE: The script will create operators of the nodes defined above and
# add them as genesis accounts in addition to the ones defined below. 
GENESIS_ADDRESSES=(
    "seda..."
    "seda..."
)
