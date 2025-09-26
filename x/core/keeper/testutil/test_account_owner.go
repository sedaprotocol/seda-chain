package testutil

import (
	"github.com/sedaprotocol/seda-chain/x/core/types"
)

func (ta *TestAccount) GetOwner() (*types.QueryOwnerResponse, error) {
	msg := &types.QueryOwnerRequest{}
	return ta.fixture.CoreQuerier.Owner(ta.fixture.Context(), msg)
}

func (ta *TestAccount) GetPendingOwner() (*types.QueryPendingOwnerResponse, error) {
	msg := &types.QueryPendingOwnerRequest{}
	return ta.fixture.CoreQuerier.PendingOwner(ta.fixture.Context(), msg)
}

func (ta *TestAccount) Paused() (*types.QueryPausedResponse, error) {
	msg := &types.QueryPausedRequest{}
	return ta.fixture.CoreQuerier.Paused(ta.fixture.Context(), msg)
}

func (ta *TestAccount) TransferOwnership(newOwner string) (*types.MsgTransferOwnershipResponse, error) {
	msg := &types.MsgTransferOwnership{
		Sender:   ta.Address(),
		NewOwner: newOwner,
	}
	return ta.fixture.CoreMsgServer.TransferOwnership(ta.fixture.Context(), msg)
}

func (ta *TestAccount) AcceptOwnership() (*types.MsgAcceptOwnershipResponse, error) {
	msg := &types.MsgAcceptOwnership{
		Sender: ta.Address(),
	}
	return ta.fixture.CoreMsgServer.AcceptOwnership(ta.fixture.Context(), msg)
}

func (ta *TestAccount) Pause() (*types.MsgPauseResponse, error) {
	msg := &types.MsgPause{
		Sender: ta.Address(),
	}
	return ta.fixture.CoreMsgServer.Pause(ta.fixture.Context(), msg)
}

func (ta *TestAccount) Unpause() (*types.MsgUnpauseResponse, error) {
	msg := &types.MsgUnpause{
		Sender: ta.Address(),
	}
	return ta.fixture.CoreMsgServer.Unpause(ta.fixture.Context(), msg)
}

func (ta *TestAccount) SetStakingConfig(config types.StakingConfig) (*types.MsgUpdateParamsResponse, error) {
	msg := &types.MsgUpdateParams{
		Authority: ta.Address(),
		Params: types.Params{
			StakingConfig: &config,
		},
	}
	return ta.fixture.CoreMsgServer.UpdateParams(ta.fixture.Context(), msg)
}
