
CHAIN_ID=seda-1-testnet
GENESIS_TIME="2024-01-18T22:00:00.000000Z"

CHAIN_VERSION=v0.0.4
WASMVM_VERSION=v1.5.2

LOCAL_BIN=$(git rev-parse --show-toplevel)/build/seda-chaind # chain binary executable on your machine

HOME_DIR=$HOME/.seda-chain # chain directory
HOME_CONFIG_DIR=$HOME_DIR/config # chain config directory

NODE_DIR=./$CHAIN_ID-nodes # where node directories will be created
WASM_DIR=./artifacts # where Wasm files will be downloaded

DENOM_METADATA='[
    {
        "description": "The token asset for SEDA Chain",
        "denom_units": [
            {
                "denom": "aseda",
                "exponent": 0,
                "aliases": [
                "attoseda"
                ]
            },
            {
                "denom": "seda",
                "exponent": 18,
                "aliases": []
            }
        ],
        "base": "aseda",
        "display": "seda",
        "name": "seda",
        "symbol": "SEDA"
    }
]'

IBC_ALLOWED_CLIENTS='[
    "06-solomachine",
    "07-tendermint"
]'

#######################################
########### VALIDATOR NODES ###########
#######################################
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

#######################################
########## GENESIS ACCOUNTS ###########
#######################################
# Standard genesis accounts
# NOTE: The script will create operators of the nodes defined above and
# add them as genesis accounts in addition to the ones defined below. 
GENESIS_ADDRESSES=(
    "seda..."
    "seda..."
)

SATOSHI=seda... # if set, creates a genesis account with 100x seda tokens compared to standard genesis account
FAUCET=seda... # if set, creates a genesis account with 10x seda tokens compared to standard genesis account

#######################################
######### COSMWASM CONTRACTS ##########
#######################################
CONTRACTS_VERSION=v0.0.1-rc # latest or seda-chain-contracts release version

#######################################
############### GITHUB ################
#######################################
GITHUB_TOKEN=ghp_... # github token for accessing seda-chain-contracts repo
