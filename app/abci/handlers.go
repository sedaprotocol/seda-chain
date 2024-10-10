package abci

import (
	"encoding/json"

	abci "github.com/cometbft/cometbft/abci/types"

	addresscodec "cosmossdk.io/core/address"
	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/app/utils"
	batchingtypes "github.com/sedaprotocol/seda-chain/x/batching/types"
)

type Handlers struct {
	batchingKeeper        BatchingKeeper
	pubKeyKeeper          PubKeyKeeper
	stakingKeeper         StakingKeeper
	validatorAddressCodec addresscodec.Codec
	signer                utils.SEDASigner
	logger                log.Logger
}

func NewHandlers(
	bk BatchingKeeper,
	pkk PubKeyKeeper,
	sk StakingKeeper,
	vac addresscodec.Codec,
	logger log.Logger,
) *Handlers {
	return &Handlers{
		batchingKeeper:        bk,
		pubKeyKeeper:          pkk,
		stakingKeeper:         sk,
		validatorAddressCodec: vac,
		logger:                logger,
	}
}

func (h *Handlers) SetSEDASigner(signer utils.SEDASigner) {
	h.signer = signer
}

// ExtendVoteHandler handles the ExtendVote ABCI to sign a batch created
// from the previous block.
func (h *Handlers) ExtendVoteHandler() sdk.ExtendVoteHandler {
	return func(ctx sdk.Context, _ *abci.RequestExtendVote) (*abci.ResponseExtendVote, error) {
		h.logger.Debug("start extend vote handler")

		// Sign the batch created from the last block's end blocker.
		batch, err := h.batchingKeeper.GetBatchForHeight(ctx, ctx.BlockHeight()-1)
		if err != nil {
			return nil, err
		}
		signature, err := h.signer.Sign(batch.BatchId, utils.Secp256k1)
		if err != nil {
			return nil, err
		}

		h.logger.Debug("submitting batch signature", "signature", signature)
		return &abci.ResponseExtendVote{VoteExtension: signature}, nil
	}
}

// VerifyVoteExtensionHandler handles the VerifyVoteExtension ABCI to
// verify the batch signature included in the pre-commit vote against
// the public key registered in the pubkey module.
func (h *Handlers) VerifyVoteExtensionHandler() sdk.VerifyVoteExtensionHandler {
	return func(ctx sdk.Context, req *abci.RequestVerifyVoteExtension) (*abci.ResponseVerifyVoteExtension, error) {
		h.logger.Debug("start verify vote extension handler", "request", req)

		batch, err := h.batchingKeeper.GetBatchForHeight(ctx, ctx.BlockHeight()-1)
		if err != nil {
			return nil, err
		}
		ok, err := h.verifyBatchSignature(ctx, batch.BatchId, req.VoteExtension[:64], req.ValidatorAddress, utils.Secp256k1)
		if !ok || err != nil {
			return nil, err
		}

		h.logger.Debug("successfully verified signature", "request", req.ValidatorAddress, "height", req.Height)
		return &abci.ResponseVerifyVoteExtension{Status: abci.ResponseVerifyVoteExtension_ACCEPT}, nil
	}
}

// PrepareProposalHandler handles the PrepareProposal ABCI to inject
// a canonical set of vote extensions in the proposal.
func (h *Handlers) PrepareProposalHandler() sdk.PrepareProposalHandler {
	return func(ctx sdk.Context, req *abci.RequestPrepareProposal) (*abci.ResponsePrepareProposal, error) {
		h.logger.Debug("start prepare proposal handler")

		err := baseapp.ValidateVoteExtensions(ctx, h.stakingKeeper, req.Height, ctx.ChainID(), req.LocalLastCommit)
		if err != nil {
			return nil, err
		}

		proposalTxs := req.Txs
		if req.Height > ctx.ConsensusParams().Abci.VoteExtensionsEnableHeight {
			bz, err := json.Marshal(req.LocalLastCommit)
			if err != nil {
				h.logger.Error("failed to marshal extended votes", "err", err)
			}
			proposalTxs = append([][]byte{bz}, proposalTxs...)
		}

		return &abci.ResponsePrepareProposal{
			Txs: proposalTxs,
		}, nil
	}
}

// ProcessProposalHandler handles the ProcessProposal ABCI to validate
// the canonical set of vote extensions injected by the proposer.
func (h *Handlers) ProcessProposalHandler() sdk.ProcessProposalHandler {
	return func(ctx sdk.Context, req *abci.RequestProcessProposal) (*abci.ResponseProcessProposal, error) {
		if req.Height > ctx.ConsensusParams().Abci.VoteExtensionsEnableHeight {
			var extendedVotes abci.ExtendedCommitInfo
			if err := json.Unmarshal(req.Txs[0], &extendedVotes); err != nil {
				h.logger.Error("failed to decode injected extended votes tx", "err", err)
				return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, nil
			}

			// Validate signatures and perform 2/3 voting power check.
			err := baseapp.ValidateVoteExtensions(ctx, h.stakingKeeper, req.Height, ctx.ChainID(), extendedVotes)
			if err != nil {
				return nil, err
			}

			// Verify every batch signature.
			batch, err := h.batchingKeeper.GetBatchForHeight(ctx, ctx.BlockHeight()-2)
			if err != nil {
				return nil, err
			}
			for _, vote := range extendedVotes.Votes {
				ok, err := h.verifyBatchSignature(ctx, batch.BatchId, vote.VoteExtension[:64], vote.Validator.Address, utils.Secp256k1)
				if !ok || err != nil {
					return nil, err
				}
			}
		}

		return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_ACCEPT}, nil
	}
}

// PreBlocker runs before BeginBlocker to extract the batch signatures
// from the canonical set of vote extensions injected by the proposer
// and store them.
func (h *Handlers) PreBlocker() sdk.PreBlocker {
	return func(ctx sdk.Context, req *abci.RequestFinalizeBlock) (res *sdk.ResponsePreBlock, err error) {
		// Use defer to prevent returning an error, which would cause
		// the chain to halt.
		defer func() {
			// Handle a panic.
			if r := recover(); r != nil {
				h.logger.Error("recovered from panic in pre-blocker", "err", r)
			}
			// Handle an error.
			if err != nil {
				h.logger.Error("error in pre-blocker", "err", err)
			}
			err = nil
		}()

		res = new(sdk.ResponsePreBlock)
		if req.Height > ctx.ConsensusParams().Abci.VoteExtensionsEnableHeight {
			h.logger.Debug("begin pre-block logic for batch signature persistence")

			var extendedVotes abci.ExtendedCommitInfo
			if err := json.Unmarshal(req.Txs[0], &extendedVotes); err != nil {
				h.logger.Error("failed to decode injected extended votes tx", "err", err)
				return nil, err
			}

			// Get batch number of the batch from two blocks ago.
			batch, err := h.batchingKeeper.GetBatchForHeight(ctx, ctx.BlockHeight()-2)
			if err != nil {
				return nil, err
			}
			batchNum := batch.BatchNumber

			for _, vote := range extendedVotes.Votes {
				valAddr, err := h.validatorAddressCodec.BytesToString(vote.Validator.Address)
				if err != nil {
					return nil, err
				}
				batchSigs := batchingtypes.BatchSignatures{
					ValidatorAddr: valAddr,
					VotingPower:   vote.Validator.Power,
					Signatures:    vote.VoteExtension,
				}

				err = h.batchingKeeper.SetBatchSignatures(ctx, batchNum, batchSigs)
				if err != nil {
					return nil, err
				}
				h.logger.Debug("stored batch signature", "batch_number", batchNum, "validator", valAddr)
			}
		}

		return res, nil
	}
}

func (h *Handlers) verifyBatchSignature(ctx sdk.Context, batchID, signature, consAddr []byte, keyIndex utils.SEDAKeyIndex) (bool, error) {
	validator, err := h.stakingKeeper.GetValidatorByConsAddr(ctx, consAddr)
	if err != nil {
		return false, err
	}
	valOper, err := h.validatorAddressCodec.StringToBytes(validator.OperatorAddress)
	if err != nil {
		return false, err
	}
	pubkey, err := h.pubKeyKeeper.GetValidatorKeyAtIndex(ctx, valOper, keyIndex)
	if err != nil {
		return false, err
	}
	return pubkey.VerifySignature(batchID, signature), nil
}
