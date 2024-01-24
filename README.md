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

[SEDA](https://seda.xyz) is an open-source data transmission and computation network that enables a permissionless environment for developers to deploy data feeds. It is built on top of [Cosmos SDK](https://github.com/cosmos/cosmos-sdk) and [CosmWasm](https://github.com/CosmWasm/wasmd).

To learn about SEDA protocol, please visit [docs.seda.xyz](https://docs.seda.xyz).<br>
To learn how to build a local version, please read [developing](DEVELOPING.md).<br>
To learn how to contribute, please read [contributing](CONTRIBUTING.md).

## Installation

### System Requirements

The current minimal system requirements are as follows and may be subject to change based on future improvements:

 - Quad Core or larger AMD or Intel (amd64) CPU
   - ARM CPUs (e.g., Apple M1) are discouraged at this time
 - 32GB RAM (with ample swap space)
 - 1TB SSD Storage
 - 100MBPS bidirectional internet connection

Running SEDA on lower-spec hardware is feasible, but you may encounter potential performance issues or an increased risk of crashes.

### Download From Releases

Download from our Github releases [page](https://github.com/sedaprotocol/seda-chain/releases).

### Build From Source

Please check out [developing](DEVELOPING.md).

## Running the Node

This is a guide for operating and running the node.

- Individuals aiming to connect to an [external node](#linking-to-an-external-node) with SEDA.
- Those who wish to establish their own node and/or set up the node as a validator.

`seda-chaind` is the command-line tool for interfacing, or CLI for short, with the SEDA blockchain. You can check out the installation instructions [here](#installation).

Now, you're all set to engage with the SEDA blockchain via an external node. For a rundown of commands, type `seda-chaind --help`. For in-depth info on a particular command, add the `--help` flag, for example:

```bash
seda-chaind --help 
seda-chaind query --help 
seda-chaind query bank --help 
```

### Linking to An External Node

This section is for those linking to an external node, so if you want to run commands from your local machine, or don't feel like running a node yourself you can use the `seda-chaind` binary to connect to an external node. This can be done two ways:

1. Add the `--node` flag to your CLI commands, followed by the RPC endpoint in the `https://<hostname>:<port>` format.
2. Alternatively, set a default node: `seda-chaind config set client node https://[hostname]:[port]`

When connecting externally, choose a trustworthy node operator. Unscrupulous operators might tamper with query outcomes or block transactions. The SEDA team currently supports these RPC endpoints:

- Testnet-seda-node-1: `http://18.171.36.35:26657`
- Testnet-seda-node-2: `http://13.41.125.154:26657`

### Running the Node Yourself

```
<!-- make the downloaded binary executable -->
chmod +x seda-chaind-${ARCH}
<!-- chmod +x seda-chaind-amd64 -->
<!-- chmod +x seda-chaind-arm64 -->

<!-- reset the chain -->
./seda-chaind-${ARCH} tendermint unsafe-reset-all
rm -rf ~/.seda-chain || true

<!-- create your operator key -->
./seda-chaind-${ARCH} keys add <key-name>

<!-- initialize your node and join the network (optionally with an existing key using the recover flag) -->
./seda-chaind-${ARCH} join <moniker> --network <devnet|testnet> [--recover]

<!-- start your node -->
./seda-chaind-${ARCH} start
```

### Creating a validator

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
 "pubkey": {"@type":"/cosmos.crypto.ed25519.PubKey","key":"$(./seda-chaind-${ARCH} tendermint show-validator)"},
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

Use the following command to create a validator:

```
./seda-chaind-${ARCH} tx staking create-validator validator.json --from <wallet-name> --chain-id <target-chain> --node <node-url>
```

That’s it now you can find your validator operator address using the following command, which you can advertise to receive delegations:

```
./seda-chaind-${ARCH} keys show <wallet-name> --bech val -a
```

### Running the Node with Cosmovisor

Run the node as a subprocess of Cosmovisor if you want automatic upgrading, which only requires you to place a new binary in the right location before an upgrade height.

Install Cosmovisor.

```
go install cosmossdk.io/tools/cosmovisor/cmd/cosmovisor@latest

```

Then, add these lines to your profile (maybe `.profile`, `.zprofile`, or something else) to set up environment variables:

```
echo "# Cosmovisor Setup" >> ~/.profile
echo "export DAEMON_NAME=seda-chaind" >> ~/.profile
echo "export DAEMON_HOME=$HOME/.seda-chain" >> ~/.profile
echo "export DAEMON_ALLOW_DOWNLOAD_BINARIES=false" >> ~/.profile
echo "export DAEMON_LOG_BUFFER_SIZE=512" >> ~/.profile
echo "export DAEMON_RESTART_AFTER_UPGRADE=true" >> ~/.profile
echo "export UNSAFE_SKIP_BACKUP=true" >> ~/.profile
source ~/.profile
```

Initialize Cosmovisor with the chain binary and start the node.

```
cosmovisor init seda-chaind-${ARCH}
cosmovisor run start
```

Note that for an upgrade, simply run the following command to prepare Cosmovisor with the upgrade binary before the chain reaches the upgrade height.

```
cosmovisor add-upgrade <upgrade-name> <upgrade-binary-file>
```

## License

Contents of this repository are open source under [GNU General Public License v3.0](LICENSE).
