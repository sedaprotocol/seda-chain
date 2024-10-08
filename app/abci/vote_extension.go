package abci

import (
	cometabci "github.com/cometbft/cometbft/abci/types"

	addresscodec "cosmossdk.io/core/address"
	"cosmossdk.io/log"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/app/utils"
)

type VoteExtensionHandler struct {
	batchingKeeper        BatchingKeeper
	pubKeyKeeper          PubKeyKeeper
	stakingKeeper         StakingKeeper
	validatorAddressCodec addresscodec.Codec
	signer                utils.SEDASigner
	logger                log.Logger
}

func NewVoteExtensionHandler(
	bk BatchingKeeper,
	pkk PubKeyKeeper,
	sk StakingKeeper,
	vac addresscodec.Codec,
	signer utils.SEDASigner,
	logger log.Logger,
) *VoteExtensionHandler {
	return &VoteExtensionHandler{
		batchingKeeper:        bk,
		pubKeyKeeper:          pkk,
		stakingKeeper:         sk,
		validatorAddressCodec: vac,
		signer:                signer,
		logger:                logger,
	}
}

func (h *VoteExtensionHandler) ExtendVoteHandler() sdk.ExtendVoteHandler {
	return func(ctx sdk.Context, req *cometabci.RequestExtendVote) (*cometabci.ResponseExtendVote, error) {
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
		return &cometabci.ResponseExtendVote{VoteExtension: signature}, nil
	}
}

func (h *VoteExtensionHandler) VerifyVoteExtensionHandler() sdk.VerifyVoteExtensionHandler {
	return func(ctx sdk.Context, req *cometabci.RequestVerifyVoteExtension) (*cometabci.ResponseVerifyVoteExtension, error) {
		h.logger.Debug("start verify vote extension handler", "request", req)

		// Verify signature.
		batch, err := h.batchingKeeper.GetBatchForHeight(ctx, ctx.BlockHeight()-1)
		if err != nil {
			return nil, err
		}

		validator, err := h.stakingKeeper.GetValidatorByConsAddr(ctx, req.ValidatorAddress)
		if err != nil {
			return nil, err
		}
		valOper, err := h.validatorAddressCodec.StringToBytes(validator.OperatorAddress)
		if err != nil {
			return nil, err
		}
		pubkey, err := h.pubKeyKeeper.GetValidatorKeyAtIndex(ctx, valOper, utils.Secp256k1)
		if err != nil {
			return nil, err
		}
		ok := pubkey.VerifySignature(batch.BatchId, req.VoteExtension)
		if !ok {
			return nil, err
		}

		h.logger.Debug("successfully verified signature", "request", req.ValidatorAddress, "height", req.Height)
		return &cometabci.ResponseVerifyVoteExtension{Status: cometabci.ResponseVerifyVoteExtension_ACCEPT}, nil
	}
}
