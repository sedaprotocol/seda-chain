package keeper

import (
	"context"
	"fmt"

	addresscodec "cosmossdk.io/core/address"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/sedaprotocol/seda-chain/app/utils"
	"github.com/sedaprotocol/seda-chain/x/staking/types"
)

var _ types.MsgServer = msgServer{}

type msgServer struct {
	stakingtypes.MsgServer
	pubkeyKeeper          types.PubkeyKeeper
	validatorAddressCodec addresscodec.Codec
}

func NewMsgServerImpl(sdkMsgServer stakingtypes.MsgServer, pubKeyKeeper types.PubkeyKeeper, valAddrCdc addresscodec.Codec) types.MsgServer {
	ms := &msgServer{
		MsgServer:             sdkMsgServer,
		pubkeyKeeper:          pubKeyKeeper,
		validatorAddressCodec: valAddrCdc,
	}
	return ms
}

func (m msgServer) CreateValidator(ctx context.Context, msg *stakingtypes.MsgCreateValidator) (*stakingtypes.MsgCreateValidatorResponse, error) {
	return nil, fmt.Errorf("not implemented")
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
	err = utils.ValidateSEDAPubKeys(msg.IndexedPubKeys)
	if err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrapf("invalid SEDA keys: %s", err)
	}
	err = m.pubkeyKeeper.StoreIndexedPubKeys(sdkCtx, valAddr, msg.IndexedPubKeys)
	if err != nil {
		return nil, err
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
