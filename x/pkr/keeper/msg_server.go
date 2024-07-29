package keeper

import (
	"context"

	"cosmossdk.io/collections"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/sedaprotocol/seda-chain/x/pkr/types"
)

type msgServer struct {
	Keeper
}

var _ types.MsgServer = msgServer{}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

func (m msgServer) AddKey(goCtx context.Context, msg *types.MsgAddKey) (*types.MsgAddKeyResponse, error) {
	if err := msg.Validate(); err != nil {
		return nil, err
	}

	valAddr, err := m.validatorAddressCodec.StringToBytes(msg.ValidatorAddress)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid validator address: %s", err)
	}

	// TODO check if validator is in the store

	pubKey, ok := msg.Pubkey.GetCachedValue().(cryptotypes.PubKey)
	if !ok {
		return nil, sdkerrors.ErrInvalidType.Wrapf("%T is not a cryptotypes.PubKey", pubKey)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	if err := m.Keeper.PubKeys.Set(ctx, collections.Join(valAddr, msg.Index), pubKey); err != nil {
		return nil, err
	}
	return &types.MsgAddKeyResponse{}, nil
}
