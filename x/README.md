# List of SEDA Chain Modules

The following are custom modules that play essential roles in the SEDA Protocol:

* [Batching](./batching/README.md) - For collecting data results, current validators, and their signatures for the purpose of facilitating delivery of data results settled on the SEDA Chain to another chain
* [Data Proxy](./data-proxy/README.md) - Registry of data proxy configurations
* [Public Key](./pubkey/README.md) - Registry of public keys for the SEDA Keys, which validators use to perform various signing duties within the SEDA Protocol
* [Tally](./tally/README.md) - For aggregating Oracle Program execution results reported by the Overlay Nodes, detection of outliers, and calculation of payouts.
* [Wasm Storage](./wasm-storage/README.md) - Storage of Oracle Programs, which include Execution Oracle Programs and Tally Oracle Programs


The following modules extend Cosmos SDK components to provide customized functionalities:

* [Slashing](./slashing/README.md) - For gatekeeping the active validator set from validators who have not registered required public keys
* [Staking](./staking/README.md) - Mainly for supporting validation creation with public keys for the SEDA Keys
* [Vesting](./vesting/README.md) - For supporting a custom vesting account type `ClawbackContinuousVestingAccount`
