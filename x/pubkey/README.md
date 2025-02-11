# Public Key Module

## Overview
The Data Proxy module is responsible for storing the configurations of data proxy provider
The public key module is a registry of public keys of the SEDA Keys, which validators use to perform various signing tasks required by the SEDA Protocol beyond consensus signing. In the current version, the only such task is signing of batches.

## State
```
0x00 | validator_address | SEDA_Key_index -> pubkey
0x01 | SEDA_Key_index                     -> proving_scheme
0x02                                      -> parameters
```

### Proving Schemes
Currently, the only supported proving scheme is secp256k1 at index 0. An activation process of this proving scheme will begin in the end blocker once the registration rate of its public keys reaches the parameter `ActivationThresholdPercent` (80% by default). Then the activation process will last for `ActivationBlockDelay` blocks (set to 11520, or roughly 1 day, by default), and if the public key registration rate remains above the threshold during this period, the proving scheme becomes activated. The validators who have failed to register their public key by the time the scheme is activated will be jailed. To unjail themselves in this case, they will have to register the required public key first before sending the unjail transaction (see the slashing module for further details).
