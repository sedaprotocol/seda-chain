# Staking Module

## Overview
The staking module wraps the SDK’s default staking module to accomplish the following:
- `TransferDelegation` and `TransferUnbonding` are taken from Agoric’s codebase ([here](https://github.com/agoric-labs/cosmos-sdk/blob/f42d86980ddfc07869846c391a03622cbd7e9188/x/staking/keeper/delegation.go#L701) and [here](https://github.com/agoric-labs/cosmos-sdk/blob/f42d86980ddfc07869846c391a03622cbd7e9188/x/staking/keeper/delegation.go#L979)) to support the clawback operations (see the vesting module for more details).
- Replace the default `CreateValidator` logic with the custom equivalent `CreateSEDAValidator`, which accepts public keys of the SEDA Keys. Once the secp256k1 proving scheme is activated, a public key from that curve is required.
- An invariant check to ensure that once the secp256k1 proving scheme is activated, no active validator is missing the required public key.
