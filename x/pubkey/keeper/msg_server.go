package keeper

import (
	"context"
	"encoding/hex"
	"fmt"

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

	err := msg.Validate()
	if err != nil {
		return nil, err
	}
	err = utils.ValidateSEDAKeys(msg.IndexedPubKeys)
	if err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrapf("invalid SEDA keys: %s", err)
	}

	valAddr, err := m.validatorAddressCodec.StringToBytes(msg.ValidatorAddr)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid validator address: %s", err)
	}
	_, err = m.stakingKeeper.GetValidator(ctx, valAddr)
	if err != nil {
		return nil, sdkerrors.ErrNotFound.Wrapf("validator not found %s", msg.ValidatorAddr)
	}

	for _, indPubKey := range msg.IndexedPubKeys {
		err := m.SetValidatorKeyAtIndex(ctx, valAddr, utils.SEDAKeyIndex(indPubKey.Index), indPubKey.PubKey)
		if err != nil {
			return nil, err
		}

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeAddKey,
				sdk.NewAttribute(types.AttributeValidatorAddr, msg.ValidatorAddr),
				sdk.NewAttribute(types.AttributePubKeyIndex, fmt.Sprintf("%d", indPubKey.Index)),
				sdk.NewAttribute(types.AttributePublicKey, hex.EncodeToString(indPubKey.PubKey)),
			),
		)
	}
	return &types.MsgAddKeyResponse{}, nil
}
