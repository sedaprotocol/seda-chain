# Batching Module

## Overview
The batching module collects data rseults, current validators, and their signatures for the purpose of facilitating delivery of data results settled on the SEDA Chain to another chain.

## State
```
0x00 | is_batched | data_request_id | data_request_height -> data_result
0x01 | data_request_id | data_request_height              -> batch_number
0x02                                                      -> current_batch_number
0x03 | block_height                                       -> batch
0x04 | batch_number                                       -> batch
0x05 | batch_number | validator_address                   -> validator_tree_entries
0x06 | batch_number                                       -> data_tree_entries
0x07 | batch_number | validator_address                   -> batch_signature
```

### Batches
Two merkle trees are constructed for each batch:
- *Validator tree*: The validator tree facilitates validation of batch signatures on the Prover Contract. Its leaves contain validator’s voting power in percentage and Ethereum-style address of the secp256k1 public key registered in the pubkey module.
- *Data result tree*: The leaves are data result IDs, which are hashes of data result contents. Once the root of a data result tree is trusted based on the batch signatures, the data results included in the tree become tamper-proof. Note at the root level, a data result tree is combined with the previous data result tree to create links between all data result trees. This way, an inclusion of any past data result can be proved against the most recent root, as long as the chain of past roots is provided.

## Batch Fraud Proof
The batching module accepts evidence of batch double signing, or signing of two different batches from the same batch number. If the evidence is proven to be valid, batch double signing is punished the same way as block double signing. That is, the validator who is proven to have committed batch double signing gets slashed, tombstoned, and jailed.
