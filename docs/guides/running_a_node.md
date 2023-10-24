# Seda-chaind Guide

This guide is tailored for:

- Individuals aiming to connect to an external node with Seda
- Those who wish to establish their own node
- Participants wanting to set up a Validator

Seda-chaind, often referred to as the "seda-chain daemon", is the primary command-line tool for interfacing with the Seda blockchain. It's the go-to tool for all node managers and validators.

## Setting Up Environment Variables

Once you've got Go installed, it's a good practice to set up the necessary environment variables:

```bash
# ~/.bashrc
export GOROOT=/usr/local/go
export GOPATH=$HOME/.go
export GOBIN=$GOPATH/bin
export PATH=$PATH:$GOPATH/bin:$GOROOT/bin
```

The above code is a bash configuration snippet, commonly used in Go development. It helps define the Go installation path, workspace, and where the executables reside.

## Getting Started with Seda-chaind

To get Seda-chaind up and running, first clone the Seda Hub repository, switch to the desired version (like v1.0.0), and then compile:

```bash
git clone https://github.com/seda-protocol/seda-chain.git 
cd seda-chain
git checkout [desired-version]
make build
```

This process will generate a 'seda-chaind' executable in the $GOBIN directory.

## Key Creation for Operators

Every node is equipped with three distinct private keys: operator, consensus, and node keys. Initially, you'll only require the operator key for transactions via Seda-chaind. The other keys will be discussed later.

To craft your operator key, execute:

```bash
seda-chaind keys add [chosen-key-name]
```

It's crucial to keep a backup of the mnemonics. Once your terminal session ends, there's no way to retrieve these keys.

For those with an existing seed phrase:

```bash
seda-chaind keys add [chosen-key-name] --recover 
```

## Linking to an External Node

If you'd rather not manage a node or Validator yourself, you can link to an external node using Seda-chaind. Just add the `--node` flag to your requests, followed by the RPC endpoint in the `https://<hostname>:<port>` format. Alternatively, set a default node:

```bash
seda-chaind config node https://[hostname]:[port]
```

When connecting externally, choose a trustworthy node operator. Unscrupulous operators might tamper with query outcomes or block transactions. The Seda team currently supports these RPC endpoints:

- Mainnet (seda-1): `https://rpc.sedaprotocol-domian.io:443`
- Testnet (ares-1): `https://testnet-rpc.sedaprotocol-domain.io:443`

You can also find a roster of public RPC endpoints in the Cosmos chain registry.

Now, you're all set to engage with the Seda blockchain via an external node. For a rundown of commands, type `seda-chaind --help`. For in-depth info on a particular command, add the `--help` flag, like so:

```bash
seda-chaind query --help 
seda-chaind query bank --help 
```


# Setting Up a Node

This guide is tailored for:

- Individuals aiming to establish a node with seda-chaind
- Participants wanting to set up a Validator with seda-chaind

## Initializing Your Node

Once you've installed seda-chaind, kickstart your node with the command:

```bash
seda-chaind init join <MONIKER> --network <NETWORK_ID> --recover
```
Moniker ID for your node
(Where network can be: devnet,...)


Executing this command will also generate your consensus and node keys. Additionally, it will create a `.seda-chaind` directory in your home folder, containing configuration files for your node:

```plaintext
~/.seda-chaind
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

Here's a brief overview of these files:

- `app.toml` - Defines application-specific settings, like gas prices and available API endpoints.
- `client.toml` - Configuration for the application's CLI. Here, you can set default parameters, such as a default `--chain-id`.
- `config.toml` - Contains settings for the underlying Tendermint consensus mechanism, which manages networking, P2P interactions, and consensus processes.
- `genesis.json` - Holds the initial set of transactions that determine the blockchain's state at its inception.
- `node_key.json` - Stores a private key used for node authentication in the P2P protocol.
- `priv_validator_key.json` - Contains the private key for Validators, vital for block and vote signing. Ensure you back up this file securely and keep its contents confidential.
- `priv_validator_state.json` - Maintains the current state of the private validator, including details like the validator's current round and height, as well as the most recent signed vote and block. Managed by the Tendermint consensus engine, this file helps prevent double-signing by your node.


finally you can run 
```
seda-chaind start
```

# Docker Setup

For docker the prosses is considerably essier

## Start the node

To start a SEDA chain node, an operator could run directly a docker run command as:

```bash
docker run --name <container-name> \
    --env-file seda.env \
    --volume <volume-name>:/root/.seda-chain \
    ghcr.io/sedaprotocol/node:v0.0.1-rc start
```

where `seda.env` is a dotenv file with:

```bash
MONIKER=
MNEMONIC=
KEYRING_PASSWORD=
NETWORK_ID=
NODE_ADDRESS=
```


Alternatively, the docker command can be run as:

```bash
docker run --name seda_node. \
--env 'MONIKER=' \
--env 'MNEMONIC=' \
--env 'KEYRING_PASSWORD=' \
--env 'NETWORK_ID=' \
--env 'NODE_ADDRESS=' \
--env-file seda.env \
--volume <volume-name>:/root/.seda-chain \
ghcr.io/sedaprotocol/node:v0.0.1-rc start
```


## Stop and Start the node

The docker container should be stoppable and resumed as:

```bash
docker stop seda_node

docker start seda_node
```