package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/sedaprotocol/seda-chain/x/proposer-sanity-check/types"

	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

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

func ProcessProposalHandler(
	baseHandler sdk.ProcessProposalHandler,
	txConfig client.TxConfig,
) sdk.ProcessProposalHandler {
	return func(ctx sdk.Context, req *abci.RequestProcessProposal) (resp *abci.ResponseProcessProposal, err error) {
		if req.Height <= 1 {
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, nil
		}
		if len(req.Txs) < 1 {
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT},
				fmt.Errorf("at least one tx [len sanity] should be available")
		}
		tx, err := txConfig.TxDecoder()(req.Txs[0])
		if err != nil {
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, err
		}
		msg, ok := decodeCounterTx(tx)
		if !ok {
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, fmt.Errorf(
				"could not decode tx")
		}

		if int(msg.GetCount()) != len(req.Txs) {
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, fmt.Errorf(
				"sanity check failed. msg.Count [%d] != len(req.Txs) [%d]", int(msg.GetCount()), len(req.Txs))
		}
		return baseHandler(ctx, req)
	}
}

func decodeCounterTx(tx sdk.Tx) (*types.MsgTxCount, bool) {
	msgs := tx.GetMsgs()
	if len(msgs) != 1 {
		return nil, false
	}
	msgNewSeed, ok := msgs[0].(*types.MsgTxCount)
	if !ok {
		return nil, false
	}
	return msgNewSeed, true
}
