package keeper

import (
	"context"

	addresscodec "cosmossdk.io/core/address"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/sedaprotocol/seda-chain/x/staking/types"
)

var _ types.MsgServer = msgServer{}

type msgServer struct {
	stakingtypes.MsgServer
	keeper                *Keeper
	accountKeeper         types.AccountKeeper
	randomnessKeeper      types.RandomnessKeeper
	validatorAddressCodec addresscodec.Codec
}

func NewMsgServerImpl(sdkMsgServer stakingtypes.MsgServer, keeper *Keeper, accKeeper types.AccountKeeper, randKeeper types.RandomnessKeeper) types.MsgServer {
	ms := &msgServer{
		MsgServer:             sdkMsgServer,
		keeper:                keeper,
		accountKeeper:         accKeeper,
		randomnessKeeper:      randKeeper,
		validatorAddressCodec: keeper.ValidatorAddressCodec(),
	}
	return ms
}

func (k msgServer) CreateValidatorWithVRF(ctx context.Context, msg *types.MsgCreateValidatorWithVRF) (*types.MsgCreateValidatorWithVRFResponse, error) {
	if err := msg.Validate(k.validatorAddressCodec); err != nil {
		return nil, err
	}

	// create an account based on VRF public key to send NewSeed txs when proposing blocks
	vrfPubKey, ok := msg.VrfPubkey.GetCachedValue().(cryptotypes.PubKey)
	if !ok {
		return nil, sdkerrors.ErrInvalidType.Wrapf("expected cryptotypes.PubKey, got %T", vrfPubKey)
	}

	addr := sdk.AccAddress(vrfPubKey.Address().Bytes())
	acc := k.accountKeeper.NewAccountWithAddress(ctx, addr)
	k.accountKeeper.SetAccount(ctx, acc)

	// register VRF public key to validator consensus address
	consPubKey, ok := msg.Pubkey.GetCachedValue().(cryptotypes.PubKey)
	if !ok {
		return nil, sdkerrors.ErrInvalidType.Wrapf("expected cryptotypes.PubKey, got %T", consPubKey)
	}
	err := k.randomnessKeeper.SetValidatorVRFPubKey(ctx, sdk.GetConsAddress(consPubKey).String(), vrfPubKey)
	if err != nil {
		return nil, err
	}

	sdkMsg := new(stakingtypes.MsgCreateValidator)
	sdkMsg.Description = msg.Description
	sdkMsg.Commission = msg.Commission
	sdkMsg.MinSelfDelegation = msg.MinSelfDelegation
	sdkMsg.ValidatorAddress = msg.ValidatorAddress
	sdkMsg.Pubkey = msg.Pubkey
	sdkMsg.Value = msg.Value

	_, err = k.MsgServer.CreateValidator(ctx, sdkMsg)
	if err != nil {
		return nil, err
	}

	return &types.MsgCreateValidatorWithVRFResponse{}, nil
}
