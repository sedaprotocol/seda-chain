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
