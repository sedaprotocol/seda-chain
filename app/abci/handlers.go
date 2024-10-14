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

const (
	// The block height difference from ExtendVote and VerifyVote to the
	// corresponding batch's creation height.
	voteExtensionOffset = -1
	// proposalOffset is the block height difference from PrepareProposal,
	// ProcessProposal, and PreBlock to the corresponding batch's creation
	// height.
	proposalOffset = -2
	// maxVoteExtensionLength is the maximum size of vote extension in
	// bytes.
	maxVoteExtensionLength = 64 * 5
)

type Handlers struct {
	defaultPrepareProposal sdk.PrepareProposalHandler
	defaultProcessProposal sdk.ProcessProposalHandler
	batchingKeeper         BatchingKeeper
	pubKeyKeeper           PubKeyKeeper
	stakingKeeper          StakingKeeper
	validatorAddressCodec  addresscodec.Codec
	signer                 utils.SEDASigner
	logger                 log.Logger
}

func NewHandlers(
	dph *baseapp.DefaultProposalHandler,
	bk BatchingKeeper,
	pkk PubKeyKeeper,
	sk StakingKeeper,
	vac addresscodec.Codec,
	logger log.Logger,
) *Handlers {
	return &Handlers{
		defaultPrepareProposal: dph.PrepareProposalHandler(),
		defaultProcessProposal: dph.ProcessProposalHandler(),
		batchingKeeper:         bk,
		pubKeyKeeper:           pkk,
		stakingKeeper:          sk,
		validatorAddressCodec:  vac,
		logger:                 logger,
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
		batch, err := h.batchingKeeper.GetBatchForHeight(ctx, ctx.BlockHeight()+voteExtensionOffset)
		if err != nil {
			return nil, err
		}
		signature, err := h.signer.Sign(batch.BatchId, utils.SEDAKeyIndexSecp256k1)
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

		batch, err := h.batchingKeeper.GetBatchForHeight(ctx, ctx.BlockHeight()+voteExtensionOffset)
		if err != nil {
			return nil, err
		}
		err = h.verifyBatchSignatures(ctx, batch.BatchId, req.VoteExtension, req.ValidatorAddress)
		if err != nil {
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

		var injection []byte
		if req.Height > ctx.ConsensusParams().Abci.VoteExtensionsEnableHeight {
			err := baseapp.ValidateVoteExtensions(ctx, h.stakingKeeper, req.Height, ctx.ChainID(), req.LocalLastCommit)
			if err != nil {
				return nil, err
			}

			injection, err = json.Marshal(req.LocalLastCommit)
			if err != nil {
				h.logger.Error("failed to marshal extended votes", "err", err)
				return nil, err
			}

			injectionSize := int64(len(injection))
			if injectionSize > req.MaxTxBytes {
				h.logger.Error(
					"vote extension size exceeds block size limit",
					"injection_size", injectionSize,
					"MaxTxBytes", req.MaxTxBytes,
				)
				return nil, ErrVoteExtensionInjectionTooBig
			}
			req.MaxTxBytes -= injectionSize
		}

		defaultRes, err := h.defaultPrepareProposal(ctx, req)
		if err != nil {
			h.logger.Error("failed to run default prepare proposal handler", "err", err)
			return nil, err
		}

		proposalTxs := defaultRes.Txs
		if injection != nil {
			proposalTxs = append([][]byte{injection}, proposalTxs...)
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
		if req.Height <= ctx.ConsensusParams().Abci.VoteExtensionsEnableHeight {
			return h.defaultProcessProposal(ctx, req)
		}

		var extendedVotes abci.ExtendedCommitInfo
		if err := json.Unmarshal(req.Txs[0], &extendedVotes); err != nil {
			h.logger.Error("failed to decode injected extended votes tx", "err", err)
			return nil, err
		}

		// Validate signatures and perform 2/3 voting power check.
		err := baseapp.ValidateVoteExtensions(ctx, h.stakingKeeper, req.Height, ctx.ChainID(), extendedVotes)
		if err != nil {
			return nil, err
		}

		// Verify every batch signature.
		batch, err := h.batchingKeeper.GetBatchForHeight(ctx, ctx.BlockHeight()+proposalOffset)
		if err != nil {
			return nil, err
		}
		for _, vote := range extendedVotes.Votes {
			err = h.verifyBatchSignatures(ctx, batch.BatchId, vote.VoteExtension, vote.Validator.Address)
			if err != nil {
				h.logger.Error("proposal contains an invalid vote extension", "vote", vote)
				return nil, err
			}
		}

		req.Txs = req.Txs[1:]
		return h.defaultProcessProposal(ctx, req)
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
		if req.Height <= ctx.ConsensusParams().Abci.VoteExtensionsEnableHeight {
			return res, nil
		}

		h.logger.Debug("begin pre-block logic for batch signature persistence")

		var extendedVotes abci.ExtendedCommitInfo
		if err := json.Unmarshal(req.Txs[0], &extendedVotes); err != nil {
			h.logger.Error("failed to decode injected extended votes tx", "err", err)
			return nil, err
		}

		// Get batch number of the batch from two blocks ago.
		batch, err := h.batchingKeeper.GetBatchForHeight(ctx, ctx.BlockHeight()+proposalOffset)
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

		return res, nil
	}
}

// verifyBatchSignature verifies the given signature of the batch ID
// against the validator's public key registered at the key index
// in the pubkey module. It returns an error unless the verification
// succeeds.
func (h *Handlers) verifyBatchSignatures(ctx sdk.Context, batchID, voteExtension, consAddr []byte) error {
	if len(voteExtension) > maxVoteExtensionLength {
		h.logger.Error("invalid vote extension length", "len", len(voteExtension))
		return ErrInvalidVoteExtensionLength
	}

	validator, err := h.stakingKeeper.GetValidatorByConsAddr(ctx, consAddr)
	if err != nil {
		return err
	}
	valOper, err := h.validatorAddressCodec.StringToBytes(validator.OperatorAddress)
	if err != nil {
		return err
	}

	// Verify secp256k1 public key
	pubkey, err := h.pubKeyKeeper.GetValidatorKeyAtIndex(ctx, valOper, utils.SEDAKeyIndexSecp256k1)
	if err != nil {
		return err
	}
	ok := pubkey.VerifySignature(batchID, voteExtension[:64])
	if !ok {
		return ErrInvalidBatchSignature
	}
	return nil
}
