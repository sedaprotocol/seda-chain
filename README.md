<p align="center">
  <a href="https://seda.xyz/">
    <img width="90%" alt="seda-chain" src="https://www.seda.xyz/images/footer/footer-image.png">
  </a>
</p>

<h1 align="center">
  SEDA Chain
</h1>

<!-- The line below is for once the repo has CI to show build status. -->
<!-- [![Build Status][actions-badge]][actions-url] -->
[![GitHub Stars][github-stars-badge]](https://github.com/sedaprotocol/seda-chain)
[![GitHub Contributors][github-contributors-badge]](https://github.com/sedaprotocol/seda-chain/graphs/contributors)
[![Discord chat][discord-badge]][discord-url]
[![Twitter][twitter-badge]][twitter-url]

<!-- The line below is for once the repo has CI to show build status. -->
<!-- [actions-badge]: https://github.com/sedaprotocol/seda-chain/actions/workflows/push.yml/badge.svg -->
[actions-url]: https://github.com/sedaprotocol/seda-chain/actions/workflows/push.yml+branch%3Amain
[github-stars-badge]: https://img.shields.io/github/stars/sedaprotocol/seda-chain.svg?style=flat-square&label=github%20stars
[github-contributors-badge]: https://img.shields.io/github/contributors/sedaprotocol/seda-chain.svg?style=flat-square
[discord-badge]: https://img.shields.io/discord/500028886025895936.svg?logo=discord&style=flat-square
[discord-url]: https://discord.gg/seda
[twitter-badge]: https://img.shields.io/twitter/url/https/twitter.com/SedaProtocol.svg?style=social&label=Follow%20%40SedaProtocol
[twitter-url]: https://twitter.com/SedaProtocol

This repo contains the our blockchain layer, written with the help of the [CosmosSDK](https://github.com/cosmos/cosmos-sdk).

To learn how to build a local version, please read [developing](DEVELOPING.md).
To learn how to contribute, please read [contributing](CONTRIBUTING.md).

## Dependencies

Our node currently doesn't rely on any runtime dependencies.

## Installation

There's a few different ways you could install and run the node.

### Download From Releases
To download from our Github releases [page](https://github.com/sedaprotocol/seda-chain/releases).

**NOTE**: The repo is currently private and this requires Github authentication. You'd have to make authenticated requests to the Github site. You can add the following to the flag to the `curl` command `-H "Authorization: token YOUR_TOKEN"`. To learn more please read Github's guide [here](https://docs.github.com/en/rest/overview/authenticating-to-the-rest-api?apiVersion=2022-11-28).

1. Now on your machine run: `curl -L -O https://github.com/sedaprotocol/seda-chain/releases/download/${SEDA_CHAIN_VERSION}/seda-chain-${SEDA_CHAIN_VERSION}-linux-${ARCH}`
   1. Replace `${SEDA_CHAIN_VERSION}` and `${ARCH}`` with the version and architecture you want. You can find a list of versions and supported architectures on our releases page linked above.
   2. You can also add `.tar.gz` if you want the tarball instead of the binary directly.
   3. You could also right click copy url for the file you want to download.
2. Download the checksum file: `curl -L -O https://github.com/sedaprotocol/seda-chain/releases/download/${SEDA_CHAIN_VERSION}/sha256sum.txt`
3. Check the checksum of the file. `sha256sum --check --ignore-missing sha256sum.txt`
   1. Should output something like: `seda-chain-${SEDA_CHAIN_VERSION}-linux-${ARCH}.tar.gz: OK`
4. You'd then want to add the binary to your path.

### Build From Source
Please check out [developing](DEVELOPING.md).

### Docker
If you have docker installed you can download the container from the Github container registry found [here](https://github.com/sedaprotocol/seda-chain/pkgs/container/node).

**NOTE**: The repo is currently private and this requires Github authentication. To learn about this process please go [here](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry).

#### Pulling the Container.
Steps:
1. `docker pull ghcr.io/sedaprotocol/node:latest`
   - Or view above link for other tags.

## Running the Node

This is a guide for operating and running the node.

- Individuals aiming to connect to an [external node](#linking-to-an-external-node) with Seda.
- Those who wish to establish their own node and/or set up the node as a validator.
  - [Without Docker](#running-the-node-yourself)
  - [With Docker](#running-the-node-yourself-dockerized)

`seda-chaind`, is the command-line tool for interfacing, or CLI for short, with the Seda blockchain. You can check out the installation instructions [here](#installation).

Now, you're all set to engage with the Seda blockchain via an external node. For a rundown of commands, type `seda-chaind --help`. For in-depth info on a particular command, add the `--help` flag, for example:

```bash
seda-chaind --help 
seda-chaind query --help 
seda-chaind query bank --help 
```

### Linking to An External Node

This section is for those linking to an external node, so if you want to run commands from your local machine, or don't feel like running a node yourself you can use the `seda-chaind` binary to connect to an external node. This can be done two ways:

1. Add the `--node` flag to your CLI commands, followed by the RPC endpoint in the `https://<hostname>:<port>` format.
2. Alternatively, set a default node: `seda-chaind config node https://[hostname]:[port]`

**NOTE**:
When connecting externally, choose a trustworthy node operator. Unscrupulous operators might tamper with query outcomes or block transactions. The Seda team currently supports these RPC endpoints:

- Testnet-seda-node-1: `http://3.10.185.200:26657`
- Testnet-seda-node-2: `http://35.179.10.147:26657`

### Running the Node Yourself
How to run the node without `docker` coming soon.

### Running the Node Yourself Dockerized

For instructions how to run the node yourself as a normal node or a validator in [docker](https://www.docker.com/). 

To pull the `docker` image check [here](#pulling-the-container).

This section will go over:
- [docker commands](#docker-commands)
- Setting up [env variables](#env-variables-configuration) for `docker`
<!-- - How to create a [key](#key-creation) -->

<!-- We recommend looking at the commands section first so if you need to run `seda-chaind` commands to create a key you can do so from within docker. -->

#### Docker Commands

Here's `docker` commands to show running the node, and executing commands so you can generate a key, and become a validator.

##### Running 

To start a SEDA chain node with `docker`

```bash
docker run -d --name seda_node \
--env-file seda.env \
--volume ~/.seda-chain:/root/.seda-chain \
--volume $(pwd)/seda.env:/seda-chain/.env \
ghcr.io/sedaprotocol/node:latest start
```

where `seda.env` is a `dotenv` described [here](#env-variables-configuration).

##### Stop and Start

The docker container should be stoppable and resumed as:

```bash
docker stop seda_node

# This runs it as a background process handled by docker.
docker start seda_node
```

##### Checking Logs

You can check the logs of your dockerized node by using the `docker logs` command. For example, to display the last 100 log lines:

```bash
docker logs seda_node -n 100
```

##### Executing `seda-chaind` Commands in Docker

Sometimes you may need to execute commands for example:
<!-- - To generate a [key](#key-creation) -->
- Check if the dockerized node is a [validator](#checking-validator-status).
- Or how to [stake](#staking)/[unstake](#unstaking).

These commands will all start with `docker exec seda_node`.

#### Env Variables Configuration

For the dockerized node to run it requires several [Environment Variables](https://wiki.archlinux.org/title/environment_variables) to be setup.

We have an example `.env` file [`.env.example`](./.env.example) file you can look at. It also describes what the env variables do.

<!-- #### Key Creation

The `docker` image will handle this for you. If in your passed in env file the mnemonic is empty, i.e. `MNEMONIC=`, it will generate one for you and update your file.

Otherwise simply have that field filled out, and it will add the account automatically. -->

#### Checking Chain Status

There are a few things you may like to check as an operator(these assume your container is running):

1. More commands coming soon.

#### Staking

**NOTE**: This assumes you already have funds in your account. If you don't please add some funds to your account before trying.

To stake run the following command:
**NOTE**: The amount at the end is the amount of tokens in `aseda` to stake.

```bash
docker exec seda_node /bin/bash -c "./staking.sh 1000"

# Which should produce some output:
gas estimate: 181315
code: 0
codespace: ""
data: ""
events: []
gas_used: "0"
gas_wanted: "0"
height: "0"
info: ""
logs: []
raw_log: '[]'
timestamp: ""
tx: null
txhash: 6C8A6C1925F3B373BBEA4DF609D8F1FAE6CDA094586763652557B527E88893A6
```

#### Unstaking

**NOTE**: This assumes you have already staked.

To unstake run the following command:
**NOTE**: The amount at the end is the amount of tokens in `aseda` to stake.
```bash
docker exec seda_node /bin/bash -c "./unstaking.sh 1000"

# Which should produce some output:
gas estimate: 164451
code: 0
codespace: ""
data: ""
events: []
gas_used: "0"
gas_wanted: "0"
height: "0"
info: ""
logs: []
txhash: 1BA768C240B379E7BBFF74D68148E95A64BFB167497C341842F4C2AF94376A77
```

#### Checking Validator Status

You can then monitor your validator status by running:
**NOTE**: To be a validator you must be one of the top 100 stakers.

```bash
docker exec seda_node /bin/bash -c "./check_validator.sh"

# Where you should see some output if error'd(which could mean it needs more time):
Error: rpc error: code = NotFound desc = rpc error: code = NotFound desc = validator sedavaloper1xd04svzj6zj93g4eknhp6aq2yyptagccp5pzst 

# Where you should see some output if successful:
commission:
  commission_rates:
    max_change_rate: "0.010000000000000000"
    max_rate: "0.200000000000000000"
    rate: "0.100000000000000000"
  update_time: "2023-11-02T22:24:25.799035935Z"
consensus_pubkey:
  '@type': /cosmos.crypto.ed25519.PubKey
  key: 0NyJ3YpZtJogW09gxxWRhnD19kYVoKpweSGmcMW+YrY=
delegator_shares: "1000.000000000000000000"
description:
  details: ""
  identity: ""
  moniker: bar
  security_contact: ""
  website: ""
jailed: false
min_self_delegation: "1"
operator_address: sedavaloper1xd04svzj6zj93g4eknhp6aq2yyptagccp5pzst
status: BOND_STATUS_UNBONDED
tokens: "1000"
unbonding_height: "0"
unbonding_ids: []
unbonding_on_hold_ref_count: "0"
unbonding_time: "1970-01-01T00:00:00Z"
```

## License

Contents of this repository are open source under [GNU General Public License v3.0](LICENSE).