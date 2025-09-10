package testutil

import (
	"github.com/sedaprotocol/seda-chain/x/core/types"
)

func (ta *TestAccount) AddToAllowlist(publicKey string) (*types.MsgAddToAllowlistResponse, error) {
	msg := &types.MsgAddToAllowlist{
		Sender:    ta.fixture.Creator.Address(),
		PublicKey: publicKey,
	}
	return ta.fixture.CoreMsgServer.AddToAllowlist(ta.fixture.Context(), msg)
}

// func (ta *TestAccount) GetAllowlist(publicKey string) (*types.QueryAllowlistResponse, error) {
// 	msg := &types.QueryAllowlistRequest{}
// 	return ta.fixture.coreQuerier.Allowlist(ta.fixture.Context(), msg)
// }

// func (ta *TestAccount) RemoveFromAllowlist(publicKey string) (*types.MsgRemoveFromAllowlistResponse, error) {
// 	msg := &types.MsgRemoveFromAllowlist{
// 		Sender:    ta.fixture.Creator.Address().String(),
// 		PublicKey: publicKey,
// 	}
// 	return ta.fixture.coreMsgServer.RemoveFromAllowlist(ta.fixture.Context(), msg)
// }
