
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
IPS=(
    "35.176.119.240"
    "18.130.188.57"
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
GENESIS_ADDRESSES=(
    "seda19gqrkdjhju0txurteag8vle90p09a5r5dd78rp"
    "seda1gnes565n2vhldm2eerm5fcuwz2mpcadvqnvped"
    "seda1wr0la8asy5wg9ja83rvdy36cmp4qrztypytdl7"
    "seda154aany5fudkp9mncekupm3hwr7w3da3dv79c4k"
    "seda15yfxudv7ek8m6ecxt4u9v5a677yhm3d662z3fg"
    "seda1c3czshqflpxs9eyns9r906gk9s9xfcpsf7rcac"
    "seda1uvraznfum5zc2tke5vu3hcj9n7a4ndcv533gnr"
    "seda1z3ecw3k2asd5gd82v7m78y6u5y5vm7xnp46lf2"
)
