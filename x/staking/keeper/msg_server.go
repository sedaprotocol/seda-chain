package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/sedaprotocol/seda-chain/app/utils"
	"github.com/sedaprotocol/seda-chain/x/staking/types"
)

// StakingMsgServer is the full staking module msg server that combines
// SDK and SEDA msg servers.
type StakingMsgServer interface {
	stakingtypes.MsgServer
	types.MsgServer
}

var (
	_ types.MsgServer  = msgServer{}
	_ StakingMsgServer = msgServer{}
)

type msgServer struct {
	stakingtypes.MsgServer
	*Keeper
}

func NewMsgServerImpl(sdkMsgServer stakingtypes.MsgServer, keeper *Keeper) StakingMsgServer {
	ms := &msgServer{
		MsgServer: sdkMsgServer,
		Keeper:    keeper,
	}
	return ms
}

// CreateValidator overrides the default CreateValidator method.
func (m msgServer) CreateValidator(_ context.Context, _ *stakingtypes.MsgCreateValidator) (*stakingtypes.MsgCreateValidatorResponse, error) {
	return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "not implemented")
}

// CreateSEDAValidator stores given SEDA public keys, if provided, and
// creates a validator.
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
