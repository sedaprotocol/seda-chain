package keeper

import (
	"context"

	addresscodec "cosmossdk.io/core/address"
	errorsmod "cosmossdk.io/errors"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/sedaprotocol/seda-chain/x/staking/types"
)

var _ types.MsgServer = msgServer{}

type msgServer struct {
	stakingtypes.MsgServer
	stakingKeeper         *stakingkeeper.Keeper
	accountKeeper         types.AccountKeeper
	randomnessKeeper      types.RandomnessKeeper
	validatorAddressCodec addresscodec.Codec
}

func NewMsgServerImpl(keeper *stakingkeeper.Keeper, accKeeper types.AccountKeeper, randKeeper types.RandomnessKeeper) types.MsgServer {
	ms := &msgServer{
		MsgServer:             stakingkeeper.NewMsgServerImpl(keeper),
		stakingKeeper:         keeper,
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
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidType, "Expecting cryptotypes.PubKey, got %T", vrfPubKey)
	}

	addr := sdk.AccAddress(vrfPubKey.Address().Bytes())
	acc := k.accountKeeper.NewAccountWithAddress(ctx, addr)
	k.accountKeeper.SetAccount(ctx, acc)

	// register VRF public key to validator consensus address
	pubKey, ok := msg.Pubkey.GetCachedValue().(cryptotypes.PubKey)
	if !ok {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidType, "Expecting cryptotypes.PubKey, got %T", pubKey)
	}
	k.randomnessKeeper.SetValidatorVRFPubKey(ctx, sdk.GetConsAddress(pubKey).String(), vrfPubKey)

	sdkMsg := new(stakingtypes.MsgCreateValidator)
	sdkMsg.Description = msg.Description
	sdkMsg.Commission = msg.Commission
	sdkMsg.MinSelfDelegation = msg.MinSelfDelegation
	sdkMsg.ValidatorAddress = msg.ValidatorAddress
	sdkMsg.Pubkey = msg.Pubkey
	sdkMsg.Value = msg.Value

	k.MsgServer.CreateValidator(ctx, sdkMsg)

	return &types.MsgCreateValidatorWithVRFResponse{}, nil
}
