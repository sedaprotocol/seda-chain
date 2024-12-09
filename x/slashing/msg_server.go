package slashing

import (
	"context"

	addresscodec "cosmossdk.io/core/address"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/slashing/types"

	"github.com/sedaprotocol/seda-chain/app/utils"
)

var _ types.MsgServer = msgServer{}

type msgServer struct {
	types.MsgServer
	pubKeyKeeper          PubKeyKeeper
	validatorAddressCodec addresscodec.Codec
}

type PubKeyKeeper interface {
	HasRegisteredKey(ctx context.Context, validatorAddr sdk.ValAddress, index utils.SEDAKeyIndex) (bool, error)
}

func NewMsgServerImpl(sdkMsgServer types.MsgServer, pubKeyKeeper PubKeyKeeper, valAddrCdc addresscodec.Codec) types.MsgServer {
	ms := &msgServer{
		MsgServer:             sdkMsgServer,
		pubKeyKeeper:          pubKeyKeeper,
		validatorAddressCodec: valAddrCdc,
	}
	return ms
}

// Unjail overrides the default Unjail method to add an additional
// check for registration of required public keys.
func (m msgServer) Unjail(ctx context.Context, req *types.MsgUnjail) (*types.MsgUnjailResponse, error) {
	valAddr, err := m.validatorAddressCodec.StringToBytes(req.ValidatorAddr)
	if err != nil {
		panic(err)
	}
	registered, err := m.pubKeyKeeper.HasRegisteredKey(ctx, valAddr, utils.SEDAKeyIndexSecp256k1)
	if err != nil {
		return nil, err
	}
	if !registered {
		return nil, sdkerrors.ErrInvalidRequest.Wrap("validator has not registered required SEDA keys")
	}

	return m.MsgServer.Unjail(ctx, req)
}
