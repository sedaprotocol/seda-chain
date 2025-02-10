# Wasm Storage Module

## Overview
The Wasm Storage module is responsible for storing Oracle Programs and the address of the Core Contract.

## State
```
0x00 | oracle_program_hash -> oracle_program
0x01                       -> core_contract_address
0x02                       -> parameters
```

### Oracle Programs 
There are two kinds of Oracle Programs: Execution Oracle Programs executed by the Overlay Nodes and Tally Oracle Programs executed by the tally module to aggregate the results reported by the Overlay Nodes. These two kinds are not distinguished in this module.

### Core Contract Registry 
The Wasm Storage module also has a capacity to instantiate the Core Contract with governance authority. Upon instantiation, the module stores the contract’s address.
