package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/sedaprotocol/seda-chain/x/proposer-sanity-check/types"

	storetypes "cosmossdk.io/store/types"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// MaxUint64 is the maximum value of a uint64.
	MaxUint64 = uint64(1) << 63
)

type Keeper struct {
	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey
}

func PrepareProposalHandler(
	baseHandler sdk.PrepareProposalHandler,
	txConfig client.TxConfig,
) sdk.PrepareProposalHandler {
	return func(ctx sdk.Context, req *abci.RequestPrepareProposal) (resp *abci.ResponsePrepareProposal, err error) {
		if req.Height <= 1 {
			return &abci.ResponsePrepareProposal{Txs: req.Txs}, nil
		}

		// In the case where there is a panic, we recover here and return an empty proposal.
		defer func() {
			if rec := recover(); rec != nil {
				fmt.Println("failed to prepare proposal", "err", err)
				resp = &abci.ResponsePrepareProposal{Txs: make([][]byte, 0)}
				err = fmt.Errorf("failed to prepare proposal: %v", rec)
			}
		}()

		panic("******************* || Test || ***********************")

		// Get the max gas limit and max block size for the proposal.
		_, _ = GetBlockLimits(ctx)
		builder := txConfig.NewTxBuilder()
		if err := builder.SetMsgs(&types.MsgTxCount{Count: int32(len(req.Txs) + 1)}); err != nil {
			return nil, err
		}

		tx := builder.GetTx()
		txBytes, err := txConfig.TxEncoder()(tx)
		if err != nil {
			return nil, err
		}
		req.Txs = append([][]byte{txBytes}, req.Txs...)

		return baseHandler(ctx, req)
	}
}

// GetBlockLimits retrieves the maximum number of bytes and gas limit allowed in a block.
func GetBlockLimits(ctx sdk.Context) (int64, uint64) {
	blockParams := ctx.ConsensusParams().Block

	// If the max gas is set to 0, then the max gas limit for the block can be infinite.
	// Otherwise we use the max gas limit casted as a uint64 which is how gas limits are
	// extracted from sdk.Tx's.
	var maxGasLimit uint64
	if maxGas := blockParams.MaxGas; maxGas > 0 {
		maxGasLimit = uint64(maxGas)
	} else {
		maxGasLimit = MaxUint64
	}

	return blockParams.MaxBytes, maxGasLimit
}
