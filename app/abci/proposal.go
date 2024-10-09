package abci

import (
	"encoding/json"

	abci "github.com/cometbft/cometbft/abci/types"
	cmttypes "github.com/cometbft/cometbft/proto/tendermint/types"

	addresscodec "cosmossdk.io/core/address"
	"cosmossdk.io/log"

	sdk "github.com/cosmos/cosmos-sdk/types"

	batchingtypes "github.com/sedaprotocol/seda-chain/x/batching/types"
)

type ProposalHandler struct {
	batchingKeeper        BatchingKeeper
	validatorAddressCodec addresscodec.Codec
	logger                log.Logger
}

func NewProposalHandler(
	bk BatchingKeeper,
	vac addresscodec.Codec,
	logger log.Logger,
) *ProposalHandler {
	return &ProposalHandler{
		batchingKeeper:        bk,
		validatorAddressCodec: vac,
		logger:                logger,
	}
}

func (h *ProposalHandler) PrepareProposalHandler() sdk.PrepareProposalHandler {
	return func(ctx sdk.Context, req *abci.RequestPrepareProposal) (*abci.ResponsePrepareProposal, error) {
		h.logger.Debug("start prepare proposal handler")

		// TODO Validate?

		proposalTxs := req.Txs

		var votes []abci.ExtendedVoteInfo
		if req.Height > ctx.ConsensusParams().Abci.VoteExtensionsEnableHeight {
			for _, vote := range req.LocalLastCommit.Votes {
				if vote.BlockIdFlag == cmttypes.BlockIDFlagCommit {
					votes = append(votes, vote)
				}
			}
		}

		bz, err := json.Marshal(votes)
		if err != nil {
			h.logger.Error("failed to marshal extended votes", "err", err)
		}
		proposalTxs = append([][]byte{bz}, proposalTxs...)

		return &abci.ResponsePrepareProposal{
			Txs: proposalTxs,
		}, nil
	}
}

func (h *ProposalHandler) ProcessProposalHandler() sdk.ProcessProposalHandler {
	return func(ctx sdk.Context, req *abci.RequestProcessProposal) (*abci.ResponseProcessProposal, error) {
		if req.Height > ctx.ConsensusParams().Abci.VoteExtensionsEnableHeight {
			var injectedVotes []abci.ExtendedVoteInfo
			if err := json.Unmarshal(req.Txs[0], &injectedVotes); err != nil {
				h.logger.Error("failed to decode injected vote extension tx", "err", err)
				return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, nil
			}

			// TODO ensure that more than 2/3 of voting power has signed the batch
		}

		return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_ACCEPT}, nil
	}
}

func (h *ProposalHandler) PreBlocker() sdk.PreBlocker {
	return func(ctx sdk.Context, req *abci.RequestFinalizeBlock) (*sdk.ResponsePreBlock, error) {
		// TODO error handling

		if req.Height > ctx.ConsensusParams().Abci.VoteExtensionsEnableHeight {
			var injectedVotes []abci.ExtendedVoteInfo
			if err := json.Unmarshal(req.Txs[0], &injectedVotes); err != nil {
				h.logger.Error("failed to decode injected vote extension tx", "err", err)
				return nil, err
			}

			batchNum := uint64(ctx.BlockHeight()) - 2 // TODO

			for _, vote := range injectedVotes {
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
			}
		}
		return &sdk.ResponsePreBlock{}, nil
	}
}
