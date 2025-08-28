package keeper

import (
	"context"
	"encoding/hex"
	"errors"
	"strconv"

	"cosmossdk.io/collections"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/sedaprotocol/seda-chain/x/sophon/types"
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

func (m msgServer) RegisterSophon(goCtx context.Context, msg *types.MsgRegisterSophon) (*types.MsgRegisterSophonResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if m.GetAuthority() != msg.Authority {
		return nil, sdkerrors.ErrorInvalidSigner.Wrapf("unauthorized authority; expected %s, got %s", m.GetAuthority(), msg.Authority)
	}

	pubKeyBytes, err := hex.DecodeString(msg.PublicKey)
	if err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrapf("invalid hex in pubkey: %s", msg.PublicKey)
	}

	hasSophon, err := m.HasSophonInfo(ctx, pubKeyBytes)
	if err != nil {
		return nil, err
	}
	if hasSophon {
		return nil, types.ErrAlreadyExists
	}

	sophonInputs := types.SophonInputs{
		OwnerAddress: msg.OwnerAddress,
		AdminAddress: msg.AdminAddress,
		Address:      msg.Address,
		PublicKey:    pubKeyBytes,
		Memo:         msg.Memo,
		Balance:      math.NewInt(0),
		UsedCredits:  math.NewInt(0),
	}

	sophonInfo, err := m.Keeper.CreateSophonInfo(ctx, pubKeyBytes, sophonInputs)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(createSophonEvent(types.EventTypeRegisterSophon, sophonInfo))

	return &types.MsgRegisterSophonResponse{}, nil
}

func (m msgServer) EditSophon(goCtx context.Context, msg *types.MsgEditSophon) (*types.MsgEditSophonResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	pubKeyBytes, err := hex.DecodeString(msg.SophonPublicKey)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "invalid hex in pubkey: %s", msg.SophonPublicKey)
	}

	sophonInfo, err := m.GetSophonInfo(ctx, pubKeyBytes)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, sdkerrors.ErrNotFound.Wrapf("sophon not found for %s", msg.SophonPublicKey)
		}
		return nil, err
	}

	if sophonInfo.OwnerAddress != msg.OwnerAddress {
		return nil, sdkerrors.ErrorInvalidSigner.Wrapf("unauthorized owner; expected %s, got %s", sophonInfo.OwnerAddress, msg.OwnerAddress)
	}

	// Update fields if provided
	if msg.NewAdminAddress != types.DoNotModifyField {
		sophonInfo.AdminAddress = msg.NewAdminAddress
	}

	if msg.NewAddress != types.DoNotModifyField {
		sophonInfo.Address = msg.NewAddress
	}

	if msg.NewMemo != types.DoNotModifyField {
		sophonInfo.Memo = msg.NewMemo
	}

	if msg.NewPublicKey != types.DoNotModifyField {
		pubKeyBytes, err := hex.DecodeString(msg.NewPublicKey)
		if err != nil {
			return nil, errorsmod.Wrapf(err, "invalid hex in new public key: %s", msg.NewPublicKey)
		}

		// Delete old storage
		err = m.DeleteSophonInfo(ctx, sophonInfo.PublicKey)
		if err != nil {
			return nil, err
		}

		// Set new public key
		sophonInfo.PublicKey = pubKeyBytes
	}

	err = m.SetSophonInfo(ctx, sophonInfo.PublicKey, sophonInfo)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(createSophonEvent(types.EventTypeUpdateSophon, sophonInfo))

	return &types.MsgEditSophonResponse{}, nil
}

func (m msgServer) AddUser(_ context.Context, _ *types.MsgAddUser) (*types.MsgAddUserResponse, error) {
	panic("not implemented")
}

func (m msgServer) TopUpUser(_ context.Context, _ *types.MsgTopUpUser) (*types.MsgTopUpUserResponse, error) {
	panic("not implemented")
}

func (m msgServer) ExpireCredits(_ context.Context, _ *types.MsgExpireCredits) (*types.MsgExpireCreditsResponse, error) {
	panic("not implemented")
}

func (m msgServer) SettleCredits(_ context.Context, _ *types.MsgSettleCredits) (*types.MsgSettleCreditsResponse, error) {
	panic("not implemented")
}

func (m msgServer) SubmitReports(_ context.Context, _ *types.MsgSubmitReports) (*types.MsgSubmitReportsResponse, error) {
	panic("not implemented")
}

func (m msgServer) UpdateParams(goCtx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid authority address: %s", msg.Authority)
	}
	if m.GetAuthority() != msg.Authority {
		return nil, sdkerrors.ErrorInvalidSigner.Wrapf("unauthorized authority; expected %s, got %s", m.GetAuthority(), msg.Authority)
	}

	if err := msg.Params.ValidateBasic(); err != nil {
		return nil, err
	}
	if err := m.SetParams(ctx, msg.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}

func createSophonEvent(eventType string, sophonInfo types.SophonInfo) sdk.Event {
	return sdk.NewEvent(
		eventType,
		sdk.NewAttribute(types.AttributeSophonPubKey, hex.EncodeToString(sophonInfo.PublicKey)),
		sdk.NewAttribute(types.AttributeSophonID, strconv.FormatUint(sophonInfo.Id, 10)),
		sdk.NewAttribute(types.AttributeOwnerAddress, sophonInfo.OwnerAddress),
		sdk.NewAttribute(types.AttributeAdminAddress, sophonInfo.AdminAddress),
		sdk.NewAttribute(types.AttributeAddress, sophonInfo.Address),
		sdk.NewAttribute(types.AttributeMemo, sophonInfo.Memo),
		sdk.NewAttribute(types.AttributeBalance, sophonInfo.Balance.String()),
		sdk.NewAttribute(types.AttributeUsedCredits, sophonInfo.UsedCredits.String()),
	)
}
