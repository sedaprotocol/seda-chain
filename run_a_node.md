# Running a Node

This section is for:

* Users who want to connect to a remote node.
* Users who want to run a node.
* Users who want to run a Validator .

## Configure Environment Variables

After installing Go, it is recommended to configure related environment variables:

```
# ~/.bashrc
export GOROOT=/usr/local/go
export GOPATH=$HOME/.go
export GOBIN=$GOPATH/bin
export PATH=$PATH:$GOPATH/bin:$GOROOT/bin
```

The provided code block is a set of environment variable assignments in a bash configuration file (~/.bashrc). These environment variables are commonly used in Go programming and their purpose is to specify the location of Go installation, workspace, and executable files:

* export GOROOT=/usr/local/go assigns the location of the Go installation directory to the GOROOT environment variable. The export keyword ensures that this variable is accessible to child processes. If using a package manager such as homebrew, this location may vary.
* export GOPATH=$HOME/.go assigns the location of the Go workspace directory to the GOPATH environment variable. The workspace is where Go source code and its dependencies are stored. By default, the workspace is located in $HOME/go but can be customized using this environment variable.
* export GOBIN=$GOPATH/bin assigns the location of the Go executable files to the GOBIN environment variable. This variable specifies the directory where Go binary files will be installed when using go install command.
* Finally, export PATH=\$PATH:\$GOPATH/bin:$GOROOT/bin adds the directories specified in GOPATH/bin and GOROOT/bin to the system's PATH variable This makes it possible to execute Go binaries from any directory in the terminal by simply typing their name.
Overall, this is a convenient way to set up the Go development environment by specifying the important directories and their locations as environment variables.

## Installing seda-chaind

seda-chaind (short for seda-chain daemon”) is the command line interface that connects with SEDA and allows you to interact with the SEDA blockchain. Every node operator and active validator uses seda-chaind to interact with their node.

To install seda-chaind, clone the seda-chain repository, checkout the latest tag, and compile the code:

git clone <git@github.com>:sedaprotocol/seda-chain.git <br>
git checkout [latest-tag] <br>
make install <br>
A seda-chaind executable will be created in the $GOBIN directory. <br>

## Generate Operator Key

Each node comes with three private keys: an operator key, a consensus key, and a node key. At this point, you only need an operator key to transact using seda-chaind (which can be achieved by connecting to a remote node). Later sections will cover consensus and node keys as well.

To generate your operator key, run the following command:

```
seda-chaind keys add <key-name>
```

Make sure you save the mnemonics! After you end your terminal session, your keys cannot be recovered.

To use an existing seed phrase, run the following command:

```
seda-chaind keys add <key-name> --recover 
```

## Connecting to a Remote Node

Users who prefer to not operate a node or Validator can connect to a remote node with seda-chaind by appending the --node flag at the end of requests along with an RPC endpoint in the https://<host>:<port> format. Alternatively, Users can configure a default node using the following command:

```
seda-chaind config node https://[host]:[port]
```

If you are connecting to a remote node, select a node operator that you can trust. Malicious operators can alter query results and censor transactions. SEDA currently maintaisn the following RPC endpoints for public use:

* Testnet (TODO): TODO
* Devnet (seda-1): TODO

At this point, you can begin interacting with the SEDA blockchain through a remote node. To learn about the list of available commands, run seda-chaind --help in your terminal. For more information about a specific command, append the --help flag at the end of your request, for example:

```
seda-chaind query --help 
seda-chaind query bank --help
```

## Joining Testnet/Devnet

### Initialize your node

```
<!-- reset the chain -->
$BIN tendermint unsafe-reset-all
rm -rf ~/.seda-chain || true

<!-- In order to join the network, a node first needs to know a few peers to connect to. A seed node is a type of node that connects to a large number of peers and informs new nodes of available peers in the network. -->
<!-- After installing seda-chaind, you can initialize your node and join the network by running the following command:  -->
<!-- This command will sync the Genesis State with your target network and set seed nodes -->
seda-chaind join <your-moniker> --network <network> [--recover]
```

Replace \<your-moniker\> with any string you’d like. This is the name to identify your server. For prospective Validators, this is NOT your validator's moniker, which we will create later.

Running this command also creates your consensus and node keys, as well as a .seda-chain folder under your home directory with some config files for your node:

```
~/.seda-chain
├─┬ config
│ ├── app.toml
│ ├── client.toml
│ ├── config.toml
│ ├── genesis.json
│ ├── node_key.json
│ ├── priv_validator_key.json
└─┬ data
  └── priv_validator_state.json
```

Let's walk over each file created:

* app.toml - Used to specify application-level settings, such as gas price, API endpoints to serve, and so on
* client.toml - The config for the app’s command line interface. This is where you can set defaul parameters, such as a default --chain-id.
* config.toml - Used to specify settings for the underlying Tendermint consensus engine, which handles networking, P2P connections, and consensus.
* genesis.json - Contains the initial set of transactions that defines the state of the blockchain at its inception.
* node_key.json - Contains a private key that is used for node authentication in the peer-to-peer (p2p) protocol. priv_validator_key.json - Contains the private key for Validators, which is used to sign blocks and votes.  You should back up this file and don't show anyone else its content.
* priv_validator_state.json - used to store the current state of the private validator. This includes information such as the current round and height of the validator, as well as the latest signed vote and block. This file is typically managed by the Tendermint consensus engine and used to prevent your node from double-signing.

### Run the Node

You can now launch the network with seda-chaind start!

However, running the network this way requires a shell to always be open. You can, instead, create a service file that will manage running the network for you.

Once you’ve started your node, you will need to wait for the node to sync up to the latest block. To check the node's sync status, you can run the following command:

```
seda-chaind status 2>&1 | jq
```

jq formats the output into a more readable format. 2>&1 is necessary due to a bug where Cosmos SDK mistakenly outputs the status to stderr instead of stdout. Your node is synced up if SyncInfo.catching_up is false.

## Running as a service

We will now run our executable as a service in order for it to be easily managed. In your system directory as a root user at /etc/systemd/system create a new service file named validator.service

```
nano validator.service
```

Use the below service file and change any specific parameters respective to your setup.

```
[Unit]
Description=SEDA Testnet Validator
After=network-online.target

[Service]
User=validator
ExecStart=/home/validator/go/bin/seda-chaind start --x-crisis-skip-assert-invariants
Restart=always
RestartSec=3
LimitNOFILE=4096

[Install]
WantedBy=multi-user.target
```

The following are systemd commands, used to manage your service:

```
<!-- Enable automatic restart of your daemon -->
systemctl enable validator.service

<!-- To reload your systemd files, run this after you have edited your service file -->
systemctl daemon-reload

<!-- Restart your service -->
systemctl restart validator.service

<!-- Start your service -->
systemctl start validator.service

<!-- Stop your service -->
systemctl stop validator.service

<!-- View logs -->
journalctl -u validator.service -f

```

## Create Validator

In order to create your validator, make sure you are fully synced to the latest block height of the network.

You can check by using the following command:

```
curl -s localhost:26657/status | jq .result | jq .sync_info
```

In the output of the above command make sure catching_up is false

```
“catching_up”: false
```

Create a validator.json file and fill in the create-validator tx parameters:

```
{
 "pubkey": {"@type":"/cosmos.crypto.ed25519.PubKey","key":"$(seda-chaind tendermint show-validator)"},
 "amount": "1000000000000000000000000000000000aseda", 
 "moniker": "the moniker for your validator",
 "identity": "optional identity signature (ex. UPort or Keybase) This key will be used by block explorers to identify the validator.",
 "website": "validator's (optional) website",
 "security": "validator's (optional) security contact email",
 "details": "validator's (optional) details",
 "commission-rate": "0.1",
 "commission-max-rate": "0.2",
 "commission-max-change-rate": "0.01",
 "min-self-delegation": "1" 
}
```

Let’s go through some flags:

* amount: The amount of aseda (the cryptocurrency used on the SEDA network) that the validator will stake as part of its candidacy.
* identity: The [Keybase](https://keybase.io/) PGP key associated with the validator's keybase.io account. This key will be used by block explorers to identify the validator.
* commission-rate: The percentage of rewards that the validator charges for its services. <br>
Note: The commission-rate value must adhere to the following rules: <br>
1- Must be between 0 and the validator's commission-max-rate. <br>
2- Must not exceed the validator's commission-max-change-rate which is maximum % point change rate per day. In other words, a validator can only change its commission once per day and within commission-max-change-rate bounds. <br>
Warning: Please note that some parameters such as commission-max-rate and commission-max-change-rate cannot be changed once your validator is up and running. <br>
* commission-max-rate: The maximum percentage that the validator can charge for its services. This number can not be changed and is meant to increase trust between you and your delegators. If you wish to go above this limit, you will have to create a new validator.
* commission-max-change-rate: The maximum percentage that the validator can increase or decrease its commission rate by.
* min-self-delegation: The minimum amount of the validator's own tokens that the validator must hold in order to remain active. If you withdraw your self-delegation below this threshold, your validator will be immediately removed from the active set. Your validator will not be slashed, but will stop earning staking rewards. This is considered the proper way for a validator to voluntarily cease operation. NOTE: If you intend to shut down your Validator, make sure to communicate with your delegators at least 14 days before withdrawing your self-delegation so that they have sufficient time to redelegate and not miss out on staking rewards.

Create a validator using the following command:

```
seda-chaind tx staking create-validator validator.json --from <wallet-name> --chain-id <target-chain> --node <node-url>
```

That’s it now you can find your validator operator address using the following command, which you can advertise to receive delegations:

```
seda-chaind keys show <wallet-name> --bech val -a
```

## Useful Commands

```
<!-- Check the maximum number of validators in the active set -->
seda-chaind q staking params | grep max_validators

<!-- Display a list of Bonded validators -->
seda-chaind q staking validators -o json --limit=1000 | jq '.validators[] | select(.status=="BOND_STATUS_BONDED")' | jq -r '.tokens + " - " + .description.moniker' | sort -gr | nl

<!-- Display a list of Unbonded validators -->
seda-chaind q staking validators -o json --limit=1000 | jq '.validators[] | select(.status=="BOND_STATUS_UNBONDED")' | jq -r '.tokens + " - " + .description.moniker' | sort -gr | nl
```

## Useful Scripts: TODO
