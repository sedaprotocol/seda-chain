package test

import (
	abci "github.com/cometbft/cometbft/abci/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type ProposalHandler struct {
	txVerifier baseapp.ProposalTxVerifier
	txSelector baseapp.TxSelector
}

func NewDefaultProposalHandler(txVerifier baseapp.ProposalTxVerifier) *ProposalHandler {
	return &ProposalHandler{
		txVerifier: txVerifier,
		txSelector: baseapp.NewDefaultTxSelector(),
	}
}

func (h *ProposalHandler) PrepareProposalHandler(txConfig client.TxConfig) sdk.PrepareProposalHandler {
	return func(ctx sdk.Context, req *abci.RequestPrepareProposal) (*abci.ResponsePrepareProposal, error) {
		//
		// fill in here
		//
		return &abci.ResponsePrepareProposal{Txs: [][]byte{}}, nil
	}
}

func (h *ProposalHandler) ProcessProposalHandler(txConfig client.TxConfig) sdk.ProcessProposalHandler {
	return func(ctx sdk.Context, req *abci.RequestProcessProposal) (*abci.ResponseProcessProposal, error) {
		//
		// fill in here
		//
		return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_ACCEPT}, nil
	}
}
