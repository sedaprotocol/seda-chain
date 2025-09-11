package testutil

import (
	"github.com/sedaprotocol/seda-chain/x/core/types"
)

func (ta *TestAccount) GetOwner() (*types.QueryOwnerResponse, error) {
	msg := &types.QueryOwnerRequest{}
	return ta.fixture.coreQuerier.Owner(ta.fixture.Context(), msg)
}

func (ta *TestAccount) GetPendingOwner() (*types.QueryPendingOwnerResponse, error) {
	msg := &types.QueryPendingOwnerRequest{}
	return ta.fixture.coreQuerier.PendingOwner(ta.fixture.Context(), msg)
}

func (ta *TestAccount) Paused() (*types.QueryPausedResponse, error) {
	msg := &types.QueryPausedRequest{}
	return ta.fixture.coreQuerier.Paused(ta.fixture.Context(), msg)
}

func (ta *TestAccount) TransferOwnership(newOwner string) (*types.MsgTransferOwnershipResponse, error) {
	msg := &types.MsgTransferOwnership{
		Sender:   ta.Address(),
		NewOwner: newOwner,
	}
	return ta.fixture.coreMsgServer.TransferOwnership(ta.fixture.Context(), msg)
}

func (ta *TestAccount) AcceptOwnership() (*types.MsgAcceptOwnershipResponse, error) {
	msg := &types.MsgAcceptOwnership{
		Sender: ta.Address(),
	}
	return ta.fixture.coreMsgServer.AcceptOwnership(ta.fixture.Context(), msg)
}

func (ta *TestAccount) Pause() (*types.MsgPauseResponse, error) {
	msg := &types.MsgPause{
		Sender: ta.Address(),
	}
	return ta.fixture.coreMsgServer.Pause(ta.fixture.Context(), msg)
}

func (ta *TestAccount) Unpause() (*types.MsgUnpauseResponse, error) {
	msg := &types.MsgUnpause{
		Sender: ta.Address(),
	}
	return ta.fixture.coreMsgServer.Unpause(ta.fixture.Context(), msg)
}
