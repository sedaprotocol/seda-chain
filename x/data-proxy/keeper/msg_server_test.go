package keeper_test

import (
	"github.com/sedaprotocol/seda-chain/x/data-proxy/types"
)

func (s *KeeperTestSuite) TestMsgServer_RegisterDataProxy() {
	tests := []struct {
		name         string
		msg          *types.MsgRegisterDataProxy
		valAddrBytes []byte
		wantErr      error
	}{
		{
			name:    "Happy path",
			msg:     &types.MsgRegisterDataProxy{},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			// TODO
		})
	}
}
