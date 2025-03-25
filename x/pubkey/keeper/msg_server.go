package keeper

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/sedaprotocol/seda-chain/app/utils"
	"github.com/sedaprotocol/seda-chain/x/pubkey/types"
)

type msgServer struct {
	*Keeper
}

var _ types.MsgServer = msgServer{}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper *Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

func (m msgServer) AddKey(goCtx context.Context, msg *types.MsgAddKey) (*types.MsgAddKeyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate the message.
	err := msg.ValidateBasic()
	if err != nil {
		return nil, err
	}
	err = utils.ValidateSEDAPubKeys(msg.IndexedPubKeys)
	if err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrapf("invalid SEDA keys: %s", err)
	}

	// Verify that the validator exists.
	valAddr, err := m.validatorAddressCodec.StringToBytes(msg.ValidatorAddr)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid validator address: %s", err)
	}
	_, err = m.stakingKeeper.GetValidator(ctx, valAddr)
	if err != nil {
		return nil, sdkerrors.ErrNotFound.Wrapf("validator not found %s", msg.ValidatorAddr)
	}

	// Retrieve previously registered keys.
	previousKeys, err := m.GetValidatorKeys(ctx, msg.ValidatorAddr)
	if err != nil && !errors.Is(err, sdkerrors.ErrNotFound) {
		return nil, err
	}

	// Store the public keys.
	err = m.StoreIndexedPubKeys(ctx, valAddr, msg.IndexedPubKeys)
	if err != nil {
		return nil, err
	}

	if len(previousKeys.IndexedPubKeys) > 0 {
		events := make(sdk.Events, len(previousKeys.IndexedPubKeys))
		for i, pk := range previousKeys.IndexedPubKeys {
			events[i] = sdk.NewEvent(
				types.EventTypeRemoveKey,
				sdk.NewAttribute(types.AttributeValidatorAddr, msg.ValidatorAddr),
				sdk.NewAttribute(types.AttributePubKeyIndex, fmt.Sprintf("%d", pk.Index)),
				sdk.NewAttribute(types.AttributePublicKey, hex.EncodeToString(pk.PubKey)),
			)
		}

		ctx.EventManager().EmitEvents(events)
	}

	return &types.MsgAddKeyResponse{}, nil
}

func (m msgServer) UpdateParams(goCtx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid authority address: %s", msg.Authority)
	}
	if m.GetAuthority() != msg.Authority {
		return nil, sdkerrors.ErrorInvalidSigner.Wrapf("unauthorized authority; expected %s, got %s", m.GetAuthority(), msg.Authority)
	}

	if err := msg.Params.Validate(); err != nil {
		return nil, err
	}
	if err := m.SetParams(ctx, msg.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}
