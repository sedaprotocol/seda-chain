package abci

import (
	"cosmossdk.io/log"
	cometabci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/app/utils"
	batchingkeeper "github.com/sedaprotocol/seda-chain/x/batching/keeper"
	pubkeykeeper "github.com/sedaprotocol/seda-chain/x/pubkey/keeper"
)

type VoteExtensionHandler struct {
	batchingKeeper batchingkeeper.Keeper // TODO Use expected keeper?
	pubKeyKeeper   pubkeykeeper.Keeper   // TODO Use expected keeper?
	signer         utils.SEDASigner
	logger         log.Logger
}

func NewVoteExtensionHandler(
	bk batchingkeeper.Keeper,
	pkk pubkeykeeper.Keeper,
	signer utils.SEDASigner,
	logger log.Logger,
) *VoteExtensionHandler {
	return &VoteExtensionHandler{
		batchingKeeper: bk,
		pubKeyKeeper:   pkk,
		signer:         signer,
		logger:         logger,
	}
}

func (h *VoteExtensionHandler) ExtendVoteHandler() sdk.ExtendVoteHandler {
	return func(ctx sdk.Context, req *cometabci.RequestExtendVote) (resp *cometabci.ResponseExtendVote, err error) {
		defer func() {
			// TODO error and panic handling
		}()

		h.logger.Debug("start extend vote handler")
		batch, err := h.batchingKeeper.GetCurrentBatch(ctx)
		if err != nil {
			return nil, err
		}
		if batch.BlockHeight != ctx.BlockHeight()-1 {
			return nil, ErrNoBatchForCurrentHeight
		}

		signature, err := h.signer.Sign(batch.BatchId, utils.Secp256k1)
		if err != nil {
			return nil, err
		}

		h.logger.Debug("signature", "signature", signature)

		return &cometabci.ResponseExtendVote{VoteExtension: signature}, nil
	}
}

func (h *VoteExtensionHandler) VerifyVoteExtensionHandler() sdk.VerifyVoteExtensionHandler {
	return func(ctx sdk.Context, req *cometabci.RequestVerifyVoteExtension) (_ *cometabci.ResponseVerifyVoteExtension, err error) {
		defer func() {
			// TODO error and panic handling
		}()

		h.logger.Debug("start verify vote extension handler")

		// Verify signature.
		batch, err := h.batchingKeeper.GetCurrentBatch(ctx)
		if err != nil {
			return nil, err
		}
		if batch.BlockHeight != ctx.BlockHeight()-1 {
			return nil, ErrNoBatchForCurrentHeight
		}

		pubkey, err := h.pubKeyKeeper.GetValidatorKeyAtIndex(ctx, req.ValidatorAddress, utils.Secp256k1)
		if err != nil {
			return &cometabci.ResponseVerifyVoteExtension{Status: cometabci.ResponseVerifyVoteExtension_REJECT}, err
		}
		ok := pubkey.VerifySignature(batch.BatchId, req.VoteExtension)
		if !ok {
			return &cometabci.ResponseVerifyVoteExtension{Status: cometabci.ResponseVerifyVoteExtension_REJECT}, err
		}

		return &cometabci.ResponseVerifyVoteExtension{Status: cometabci.ResponseVerifyVoteExtension_ACCEPT}, nil
	}
}
