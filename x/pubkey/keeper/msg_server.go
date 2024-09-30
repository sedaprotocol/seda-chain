package keeper

import (
	"context"
	"fmt"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/sedaprotocol/seda-chain/app/utils"
	"github.com/sedaprotocol/seda-chain/x/pubkey/types"
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
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := msg.Validate(); err != nil {
		return nil, err
	}

	valAddr, err := m.validatorAddressCodec.StringToBytes(msg.ValidatorAddr)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid validator address: %s", err)
	}
	if _, err := m.stakingKeeper.GetValidator(ctx, valAddr); err != nil {
		return nil, sdkerrors.ErrNotFound.Wrapf("validator not found %s", msg.ValidatorAddr)
	}

	for _, indPubKey := range msg.IndexedPubKeys {
		pubKey, ok := indPubKey.PubKey.GetCachedValue().(cryptotypes.PubKey)
		if !ok {
			return nil, sdkerrors.ErrInvalidType.Wrapf("%T is not a cryptotypes.PubKey", pubKey)
		}
		if err := m.SetValidatorKeyAtIndex(ctx, valAddr, utils.SEDAKeyIndex(indPubKey.Index), pubKey); err != nil {
			return nil, err
		}

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeAddKey,
				sdk.NewAttribute(types.AttributeValidatorAddr, msg.ValidatorAddr),
				sdk.NewAttribute(types.AttributePubKeyIndex, fmt.Sprintf("%d", indPubKey.Index)),
				sdk.NewAttribute(types.AttributePublicKey, pubKey.String()),
			),
		)
	}
	return &types.MsgAddKeyResponse{}, nil
}
