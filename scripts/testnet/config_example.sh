CHAIN_ID=seda-1-devnet
GENESIS_TIME="2024-04-24T16:00:00.000000Z"
CHAIN_VERSION=v0.1.3
WASMVM_VERSION=v1.5.2

WASM_PERMISSION_EVERYONE=true # true for everyone and false for mainnet configuration
SHORT_VOTING_PERIOD=true # true for 180s voting period or false for mainnet configuration

# If DOWNLOAD_FROM_RELEASE is set to false, specify RUN_NO and ARTIFACT_NO so the script
# can download the artifact.
DOWNLOAD_FROM_RELEASE=false
RUN_NO=0123
ARTIFACT_NO=0123

LOCAL_BIN=$(git rev-parse --show-toplevel)/build/sedad # chain binary executable on your machine
HOME_DIR=$HOME/.sedad # chain directory
HOME_CONFIG_DIR=$HOME_DIR/config # chain config directory
NODE_DIR=./$CHAIN_ID-nodes # where node directories will be created
WASM_DIR=./artifacts # where Wasm files will be downloaded


#######################################
########### VALIDATOR NODES ###########
#######################################
# NOTE: ONLY FILL OUT THIS PART IF YOU ARE SETTING UP VALIDATOR NODES.
# NOTE: Assumes 26656 port for p2p communication
# NOTE: Assumes user is ec2-user
# NOTE: The setup node script assumes ami-0a1ab4a3fcf997a9d
IPS=(
    "xx.xx.xx.xx"
    "xx.xx.xx.xx"
)
SSH_KEY=~/.ssh/id_rsa # or id_ed25519 
MONIKERS=(
    "SEDA-node0"
    "SEDA-node1"
)
SELF_DELEGATION_AMOUNTS=(
    "27500seda"
    "27500seda"
)

#######################################
########## GENESIS ACCOUNTS ###########
#######################################
SATOSHI=seda... # if set, creates a genesis account with 100x seda tokens compared to standard genesis account
FAUCET=seda... # if set, creates a genesis account with 10x seda tokens compared to standard genesis account

# Standard genesis accounts
# NOTE: The script will create operators of the nodes defined above and
# add them as genesis accounts in addition to the ones defined below. 
GENESIS_ADDRESSES=(
    "seda..."
    "seda..."
)

#######################################
######### COSMWASM CONTRACTS ##########
#######################################
CONTRACTS_VERSION=v0.0.1-rc # latest or seda-chain-contracts release version

#######################################
############ GROUP CONFIG #############
#######################################
GROUP_OOA_MEMBERS=./group_ooa_members.json
GROUP_SECURITY_MEMBERS=./group_security_members.json
GROUP_TREASURY_MEMBERS=./group_treasury_members.json

GROUP_OOA_POLICY=./group_ooa_policy.json
GROUP_SECURITY_POLICY=./group_security_policy.json
GROUP_TREASURY_POLICY=./group_treasury_policy.json

ADMIN_SEED="mushroom energy ..." # used for creating groups - overwritten by group policy addresses anyways

#######################################
############### GITHUB ################
#######################################
GITHUB_TOKEN=ghp_... # github token for accessing seda-chain-contracts repo
