# ABCI

The ABCI package of the SEDA Chain implements the CometBFT ABCI++ interface to support batch signing. Once a new batch is created at block height `H`, the following sequence begins:

1. `ExtendVote` at `H+1` - Batch Signing
    - Validators sign the batch using the secp256k1 signature scheme and include their signatures in their pre-commit votes.
2. `VerifyVoteExtension` at `H+1` - Batch Signature Verification
    - Upon receiving a pre-commit vote with a batch signature, the validator checks the signature against the batch in the store and the corresponding public key registered in the pubkey module. The vote is only accepted if the signature verification succeeds.
3. `PrepareProposal` at `H+2` - Injecting Vote Extensions in Proposal
    - When the proposer at block `H+1` proposes a canonical set of votes for the block `H`, it also injects the votes’ extended data in the proposal so that a canonical set of batch signatures also become available to all validators.
4. `ProcessProposal` at `H+2` - Batch Signatures Validation
    - The proposed canonical set of batch signatures is checked to ensure that more than 2/3 of voting power according to the previous block’s validator set has signed the batch.
5. `PreBlock` at `H+2` - Batch Signatures Persistence
    - It is run at the beginning of `FinalizeBlock` ABCI call to store the fully-populated batch in the batching module store.
