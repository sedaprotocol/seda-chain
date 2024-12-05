package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/sedaprotocol/seda-chain/app/utils"
	"github.com/sedaprotocol/seda-chain/x/staking/types"
)

var _ types.MsgServer = msgServer{}

type msgServer struct {
	stakingtypes.MsgServer
	*Keeper
}

func NewMsgServerImpl(sdkMsgServer stakingtypes.MsgServer, keeper *Keeper) types.MsgServer {
	ms := &msgServer{
		MsgServer: sdkMsgServer,
		Keeper:    keeper,
	}
	return ms
}

func (m msgServer) CreateSEDAValidator(ctx context.Context, msg *types.MsgCreateSEDAValidator) (*types.MsgCreateSEDAValidatorResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	err := msg.Validate(m.validatorAddressCodec)
	if err != nil {
		return nil, err
	}
	valAddr, err := m.validatorAddressCodec.StringToBytes(msg.ValidatorAddress)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid validator address: %s", err)
	}

	// Validate and store the public keys.
	activated, err := m.pubKeyKeeper.IsProvingSchemeActivated(ctx, utils.SEDAKeyIndexSecp256k1)
	if err != nil {
		return nil, err
	}
	if len(msg.IndexedPubKeys) > 0 {
		err = utils.ValidateSEDAPubKeys(msg.IndexedPubKeys)
		if err != nil {
			return nil, sdkerrors.ErrInvalidRequest.Wrapf("invalid SEDA keys: %s", err)
		}
		err = m.pubKeyKeeper.StoreIndexedPubKeys(sdkCtx, valAddr, msg.IndexedPubKeys)
		if err != nil {
			return nil, err
		}
	} else if activated {
		return nil, sdkerrors.ErrInvalidRequest.Wrap("SEDA public keys are required")
	}

	// Call the wrapped CreateValidator method.
	sdkMsg := new(stakingtypes.MsgCreateValidator)
	sdkMsg.Description = msg.Description
	sdkMsg.Commission = msg.Commission
	sdkMsg.MinSelfDelegation = msg.MinSelfDelegation
	sdkMsg.ValidatorAddress = msg.ValidatorAddress
	sdkMsg.Pubkey = msg.Pubkey
	sdkMsg.Value = msg.Value

	_, err = m.MsgServer.CreateValidator(ctx, sdkMsg)
	if err != nil {
		return nil, err
	}
	return &types.MsgCreateSEDAValidatorResponse{}, nil
}

func (m msgServer) EditValidator(ctx context.Context, msg *stakingtypes.MsgEditValidator) (*stakingtypes.MsgEditValidatorResponse, error) {
	return m.MsgServer.EditValidator(ctx, msg)
}

func (m msgServer) Delegate(ctx context.Context, msg *stakingtypes.MsgDelegate) (*stakingtypes.MsgDelegateResponse, error) {
	return m.MsgServer.Delegate(ctx, msg)
}

func (m msgServer) BeginRedelegate(ctx context.Context, msg *stakingtypes.MsgBeginRedelegate) (*stakingtypes.MsgBeginRedelegateResponse, error) {
	return m.MsgServer.BeginRedelegate(ctx, msg)
}

func (m msgServer) Undelegate(ctx context.Context, msg *stakingtypes.MsgUndelegate) (*stakingtypes.MsgUndelegateResponse, error) {
	return m.MsgServer.Undelegate(ctx, msg)
}

func (m msgServer) CancelUnbondingDelegation(ctx context.Context, msg *stakingtypes.MsgCancelUnbondingDelegation) (*stakingtypes.MsgCancelUnbondingDelegationResponse, error) {
	return m.MsgServer.CancelUnbondingDelegation(ctx, msg)
}

func (m msgServer) UpdateParams(ctx context.Context, msg *stakingtypes.MsgUpdateParams) (*stakingtypes.MsgUpdateParamsResponse, error) {
	return m.MsgServer.UpdateParams(ctx, msg)
}
