package abci

import (
	"bytes"
	"encoding/json"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/crypto"

	abcitypes "github.com/cometbft/cometbft/abci/types"
	cmttypes "github.com/cometbft/cometbft/proto/tendermint/types"

	"cosmossdk.io/collections"
	addresscodec "cosmossdk.io/core/address"
	"cosmossdk.io/log"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/app/utils"
)

const (
	// BlockOffsetSign is the block height difference between the batch
	// signing phase and the corresponding batch's creation height.
	// The signing phase consists of ExtendVote and VerifyVoteExtension.
	BlockOffsetSignPhase = -1
	// BlockOffsetCollectPhase is the block height difference between the
	// batch signature collection phase and the corresponding batch's
	// creation height. The collection phase spans PrepareProposal,
	// ProcessProposal, and PreBlock to store a canonical set of batch
	// signatures.
	BlockOffsetCollectPhase = -2

	// MinVoteExtensionLength is the minimum size of vote extension in bytes.
	MinVoteExtensionLength = 65
	// MaxVoteExtensionLength is the maximum size of vote extension in bytes.
	MaxVoteExtensionLength = 65 * 5
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
	prepareHandler sdk.PrepareProposalHandler,
	processHandler sdk.ProcessProposalHandler,
	bk BatchingKeeper,
	pkk PubKeyKeeper,
	sk StakingKeeper,
	vac addresscodec.Codec,
	signer utils.SEDASigner,
	logger log.Logger,
) *Handlers {
	return &Handlers{
		defaultPrepareProposal: prepareHandler,
		defaultProcessProposal: processHandler,
		batchingKeeper:         bk,
		pubKeyKeeper:           pkk,
		stakingKeeper:          sk,
		validatorAddressCodec:  vac,
		signer:                 signer,
		logger:                 logger,
	}
}

// ExtendVoteHandler handles the ExtendVote ABCI to sign a batch created
// from the previous block.
func (h *Handlers) ExtendVoteHandler() sdk.ExtendVoteHandler {
	return func(ctx sdk.Context, _ *abcitypes.RequestExtendVote) (*abcitypes.ResponseExtendVote, error) {
		h.logger.Debug("start extend vote handler", "height", ctx.BlockHeight())

		// Check if there is a batch to sign at this block height.
		batch, err := h.batchingKeeper.GetBatchForHeight(ctx, ctx.BlockHeight()+BlockOffsetSignPhase)
		if err != nil {
			if errors.Is(err, collections.ErrNotFound) {
				h.logger.Debug("no batch to sign", "height", ctx.BlockHeight())
				return &abcitypes.ResponseExtendVote{}, nil
			}
			return nil, err
		}

		if !h.signer.IsLoaded() {
			h.logger.Info("signer is not loaded, try reloading")
			err := h.signer.ReloadIfMismatch(nil)
			if err != nil {
				h.logger.Error("failed to load signer to sign batch", "err", err)
				return nil, err
			}
		}

		// Check if the validator was in the previous validator tree. If not,
		// this means the validator has just joined the active set, so it skips
		// signing this batch. The very first batch is signed by all validators.
		if batch.BatchNumber != collections.DefaultSequenceStart {
			_, err = h.batchingKeeper.GetValidatorTreeEntry(ctx, batch.BatchNumber-1, h.signer.GetValAddress())
			if err != nil {
				if errors.Is(err, collections.ErrNotFound) {
					h.logger.Info("validator was not in the previous validator tree - not signing the batch")
				} else {
					h.logger.Error("unexpected error while checking previous validator tree entry", "err", err)
				}
				return nil, err
			}
		}

		valKeys, err := h.pubKeyKeeper.GetValidatorKeys(ctx, h.signer.GetValAddress().String())
		if err != nil {
			return nil, err
		}

		// Sign and reload the signer if the public key has changed.
		signature, err := h.signer.Sign(batch.BatchId, utils.SEDAKeyIndexSecp256k1)
		if err != nil {
			return nil, err
		}
		err = h.signer.ReloadIfMismatch(valKeys.IndexedPubKeys)
		if err != nil {
			return nil, err
		}

		h.logger.Debug(
			"submitting batch signature",
			"signature", signature,
			"batch_number", batch.BatchNumber,
		)
		return &abcitypes.ResponseExtendVote{VoteExtension: signature}, nil
	}
}

// VerifyVoteExtensionHandler handles the VerifyVoteExtension ABCI to
// verify the batch signature included in the pre-commit vote against
// the public key registered in the pubkey module.
func (h *Handlers) VerifyVoteExtensionHandler() sdk.VerifyVoteExtensionHandler {
	return func(ctx sdk.Context, req *abcitypes.RequestVerifyVoteExtension) (*abcitypes.ResponseVerifyVoteExtension, error) {
		h.logger.Debug("start verify vote extension handler", "request", req)

		batch, err := h.batchingKeeper.GetBatchForHeight(ctx, ctx.BlockHeight()+BlockOffsetSignPhase)
		if err != nil {
			if errors.Is(err, collections.ErrNotFound) {
				if req.VoteExtension == nil {
					return &abcitypes.ResponseVerifyVoteExtension{Status: abcitypes.ResponseVerifyVoteExtension_ACCEPT}, nil
				}
				h.logger.Error(
					"received vote extension even though we're skipping batching",
					"request", req.ValidatorAddress,
					"height", req.Height,
					"vote_extension", req.VoteExtension,
				)
				return &abcitypes.ResponseVerifyVoteExtension{Status: abcitypes.ResponseVerifyVoteExtension_REJECT}, nil
			}
			return nil, err
		}

		err = h.verifyBatchSignatures(ctx, batch.BatchNumber, batch.BatchId, req.VoteExtension, req.ValidatorAddress)
		if err != nil {
			h.logger.Error("failed to verify batch signature", "req", req, "err", err)
			return &abcitypes.ResponseVerifyVoteExtension{Status: abcitypes.ResponseVerifyVoteExtension_REJECT}, err
		}

		h.logger.Debug(
			"successfully verified signature",
			"request", req.ValidatorAddress,
			"height", req.Height,
			"batch_number", batch.BatchNumber,
		)
		return &abcitypes.ResponseVerifyVoteExtension{Status: abcitypes.ResponseVerifyVoteExtension_ACCEPT}, nil
	}
}

// PrepareProposalHandler handles the PrepareProposal ABCI to inject
// a canonical set of vote extensions in the proposal.
func (h *Handlers) PrepareProposalHandler() sdk.PrepareProposalHandler {
	return func(ctx sdk.Context, req *abcitypes.RequestPrepareProposal) (*abcitypes.ResponsePrepareProposal, error) {
		// Check if there is a batch whose signatures must be collected
		// at this block height.
		var collectSigs bool
		batch, err := h.batchingKeeper.GetBatchForHeight(ctx, ctx.BlockHeight()+BlockOffsetCollectPhase)
		if err != nil {
			if !errors.Is(err, collections.ErrNotFound) {
				return nil, err
			}
		} else {
			collectSigs = true
		}

		var injection []byte
		if IsVoteExtensionsEnabled(ctx) && collectSigs {
			for i, vote := range req.LocalLastCommit.Votes {
				// Verify the signatures since they're not guaranteed to have passed VerifyVoteExtension.
				if err := h.verifyBatchSignatures(ctx, batch.BatchNumber, batch.BatchId, vote.VoteExtension, vote.Validator.Address); err != nil {
					h.logger.Info(
						"failed to validate vote extension - pruning vote",
						"err", err,
						"validator", vote.Validator.Address,
					)

					// failed to validate this vote-extension, mark it as absent in the original commit
					vote.BlockIdFlag = cmttypes.BlockIDFlagAbsent
					vote.ExtensionSignature = nil
					vote.VoteExtension = nil
					req.LocalLastCommit.Votes[i] = vote
				}
			}

			// Validate after pruning to ensure we have enough voting power.
			err := ValidateVoteExtensions(ctx, h.stakingKeeper, req.LocalLastCommit)
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
			h.logger.Debug("injected local last commit", "height", req.Height)
		}
		return &abcitypes.ResponsePrepareProposal{
			Txs: proposalTxs,
		}, nil
	}
}

// ProcessProposalHandler handles the ProcessProposal ABCI to validate
// the canonical set of vote extensions injected by the proposer.
func (h *Handlers) ProcessProposalHandler() sdk.ProcessProposalHandler {
	return func(ctx sdk.Context, req *abcitypes.RequestProcessProposal) (*abcitypes.ResponseProcessProposal, error) {
		if !IsVoteExtensionsEnabled(ctx) {
			return h.defaultProcessProposal(ctx, req)
		}

		batch, err := h.batchingKeeper.GetBatchForHeight(ctx, ctx.BlockHeight()+BlockOffsetCollectPhase)
		if err != nil {
			if errors.Is(err, collections.ErrNotFound) {
				return h.defaultProcessProposal(ctx, req)
			}
			return &abcitypes.ResponseProcessProposal{Status: abcitypes.ResponseProcessProposal_REJECT}, err
		}

		if len(req.Txs) == 0 {
			h.logger.Error("proposal does not contain extended votes injection")
			return &abcitypes.ResponseProcessProposal{Status: abcitypes.ResponseProcessProposal_REJECT}, ErrNoInjectedExtendedVotesTx
		}

		var extendedVotes abcitypes.ExtendedCommitInfo
		if err := json.Unmarshal(req.Txs[0], &extendedVotes); err != nil {
			h.logger.Error("failed to decode injected extended votes tx", "err", err)
			return &abcitypes.ResponseProcessProposal{Status: abcitypes.ResponseProcessProposal_REJECT}, err
		}

		// Validate vote extensions and batch signatures.
		err = ValidateVoteExtensions(ctx, h.stakingKeeper, extendedVotes)
		if err != nil {
			return &abcitypes.ResponseProcessProposal{Status: abcitypes.ResponseProcessProposal_REJECT}, err
		}

		for _, vote := range extendedVotes.Votes {
			// Only consider extensions with pre-commit votes.
			if vote.BlockIdFlag == cmttypes.BlockIDFlagCommit {
				err = h.verifyBatchSignatures(ctx, batch.BatchNumber, batch.BatchId, vote.VoteExtension, vote.Validator.Address)
				if err != nil {
					h.logger.Error("proposal contains an invalid vote extension", "vote", vote)
					return &abcitypes.ResponseProcessProposal{Status: abcitypes.ResponseProcessProposal_REJECT}, err
				}
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
	return func(ctx sdk.Context, req *abcitypes.RequestFinalizeBlock) (res *sdk.ResponsePreBlock, err error) {
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
		if !IsVoteExtensionsEnabled(ctx) {
			return res, nil
		}

		batch, err := h.batchingKeeper.GetBatchForHeight(ctx, ctx.BlockHeight()+BlockOffsetCollectPhase)
		if err != nil {
			if errors.Is(err, collections.ErrNotFound) {
				h.logger.Debug("no batch to collect signatures for", "height", ctx.BlockHeight())
				return res, nil
			}
			return nil, err
		}
		batchNum := batch.BatchNumber

		h.logger.Debug("begin pre-block logic for storing batch signatures", "batch_number", batchNum)

		if len(req.Txs) == 0 {
			h.logger.Error("proposal does not contain extended votes injection")
			return nil, ErrNoInjectedExtendedVotesTx
		}

		var extendedVotes abcitypes.ExtendedCommitInfo
		if err := json.Unmarshal(req.Txs[0], &extendedVotes); err != nil {
			h.logger.Error("failed to decode injected extended votes tx", "err", err)
			return nil, err
		}

		for _, vote := range extendedVotes.Votes {
			// Skip votes that are absent (possibly pruned invalid votes)
			// or have no vote extension (for new validators).
			if vote.BlockIdFlag == cmttypes.BlockIDFlagAbsent || len(vote.VoteExtension) == 0 {
				continue
			}

			validator, err := h.stakingKeeper.GetValidatorByConsAddr(ctx, vote.Validator.Address)
			if err != nil {
				return nil, err
			}
			valAddr, err := h.validatorAddressCodec.StringToBytes(validator.OperatorAddress)
			if err != nil {
				return nil, err
			}
			err = h.batchingKeeper.SetBatchSigSecp256k1(ctx, batchNum, valAddr, vote.VoteExtension)
			if err != nil {
				return nil, err
			}
			h.logger.Debug("stored batch signature", "batch_number", batchNum, "validator", validator.OperatorAddress)
		}

		return res, nil
	}
}

// verifyBatchSignature verifies the given signature of the batch ID
// against the validator's public key registered at the key index
// in the pubkey module. It returns an error unless the verification
// succeeds.
func (h *Handlers) verifyBatchSignatures(ctx sdk.Context, batchNum uint64, batchID, voteExtension, consAddr []byte) error {
	if len(voteExtension) > MaxVoteExtensionLength {
		h.logger.Error("vote extension exceeds max length", "len", len(voteExtension))
		return ErrVoteExtensionTooLong
	}

	validator, err := h.stakingKeeper.GetValidatorByConsAddr(ctx, consAddr)
	if err != nil {
		return err
	}
	valOper, err := h.validatorAddressCodec.StringToBytes(validator.OperatorAddress)
	if err != nil {
		return err
	}

	// Recover and verify secp256k1 public key.
	var expectedAddr []byte
	if batchNum == collections.DefaultSequenceStart {
		pubKey, err := h.pubKeyKeeper.GetValidatorKeyAtIndex(ctx, valOper, utils.SEDAKeyIndexSecp256k1)
		if err != nil {
			return err
		}
		expectedAddr, err = utils.PubKeyToEthAddress(pubKey)
		if err != nil {
			return err
		}
	} else {
		valEntry, err := h.batchingKeeper.GetValidatorTreeEntry(ctx, batchNum-1, valOper)
		if err != nil {
			if errors.Is(err, collections.ErrNotFound) {
				if len(voteExtension) == 0 {
					return nil
				}
				return ErrUnexpectedBatchSignature
			}
			return err
		}
		expectedAddr = valEntry.EthAddress
	}

	if len(voteExtension) < MinVoteExtensionLength {
		h.logger.Error("vote extension is too short", "len", len(voteExtension))
		return ErrVoteExtensionTooShort
	}

	signature := voteExtension[:65]
	r := new(big.Int).SetBytes(signature[:32])
	s := new(big.Int).SetBytes(signature[32:64])
	v := signature[64]

	// We require the signature to be in the low-s form, which is the default for
	// the secp256k1 library used to sign the batch ID.
	if !crypto.ValidateSignatureValues(v, r, s, true) {
		return ErrInvalidBatchSignature
	}

	sigPubKey, err := crypto.Ecrecover(batchID, signature)
	if err != nil {
		return err
	}
	sigAddr, err := utils.PubKeyToEthAddress(sigPubKey)
	if err != nil {
		return err
	}

	if !bytes.Equal(expectedAddr, sigAddr) {
		return ErrInvalidBatchSignature
	}
	return nil
}
