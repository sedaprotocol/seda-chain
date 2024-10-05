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
	return func(ctx sdk.Context, _ *cometabci.RequestExtendVote) (*cometabci.ResponseExtendVote, error) {
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

		h.logger.Debug("submitting batch signature", "signature", signature)
		return &cometabci.ResponseExtendVote{VoteExtension: signature}, nil
	}
}

func (h *VoteExtensionHandler) VerifyVoteExtensionHandler() sdk.VerifyVoteExtensionHandler {
	return func(ctx sdk.Context, req *cometabci.RequestVerifyVoteExtension) (*cometabci.ResponseVerifyVoteExtension, error) {
		h.logger.Debug("start verify vote extension handler", "request", req)

		// Verify signature.
		batch, err := h.batchingKeeper.GetCurrentBatch(ctx)
		if err != nil {
			return nil, err
		}
		if batch.BlockHeight != ctx.BlockHeight()-1 {
			return nil, ErrNoBatchForCurrentHeight
		}

		validator, err := h.stakingKeeper.GetValidatorByConsAddr(ctx, req.ValidatorAddress)
		if err != nil {
			return &cometabci.ResponseVerifyVoteExtension{Status: cometabci.ResponseVerifyVoteExtension_REJECT}, err
		}
		valOper, err := h.validatorAddressCodec.StringToBytes(validator.OperatorAddress)
		if err != nil {
			return &cometabci.ResponseVerifyVoteExtension{Status: cometabci.ResponseVerifyVoteExtension_REJECT}, err
		}
		pubkey, err := h.pubKeyKeeper.GetValidatorKeyAtIndex(ctx, valOper, utils.Secp256k1)
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
