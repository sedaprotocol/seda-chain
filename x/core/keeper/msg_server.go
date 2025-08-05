package keeper

import (
	"context"
	"encoding/hex"
	"errors"

	vrf "github.com/sedaprotocol/vrf-go"

	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/sedaprotocol/seda-chain/x/core/types"
)

var _ types.MsgServer = msgServer{}

type msgServer struct {
	Keeper
}

func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

func (m msgServer) AddToAllowlist(goCtx context.Context, msg *types.MsgAddToAllowlist) (*types.MsgAddToAllowlistResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if msg.Sender != m.GetAuthority() {
		return nil, sdkerrors.ErrUnauthorized.Wrapf("unauthorized authority; expected %s, got %s", m.GetAuthority(), msg.Sender)
	}

	exists, err := m.Allowlist.Has(ctx, msg.PublicKey)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, types.ErrAlreadyAllowlisted
	}

	err = m.Allowlist.Set(ctx, msg.PublicKey)
	if err != nil {
		return nil, err
	}

	// TODO: add event
	// Ok(Response::new().add_attribute("action", "add-to-allowlist").add_event(
	// 	Event::new("seda-contract").add_attributes([
	// 		("version", CONTRACT_VERSION.to_string()),
	// 		("identity", self.public_key),
	// 		("action", "allowlist-add".to_string()),
	// 	]),
	// ))
	return &types.MsgAddToAllowlistResponse{}, nil
}

func (m msgServer) Stake(goCtx context.Context, msg *types.MsgStake) (*types.MsgStakeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Verify stake proof.
	var sequenceNum uint64
	var isExistingStaker bool
	staker, err := m.Stakers.Get(ctx, msg.PublicKey)
	if err != nil {
		if !errors.Is(err, collections.ErrNotFound) {
			return nil, err
		}
	} else {
		sequenceNum = staker.SequenceNum
		isExistingStaker = true
	}

	hash, err := msg.ComputeStakeHash("", ctx.ChainID(), sequenceNum)
	if err != nil {
		return nil, err
	}
	publicKey, err := hex.DecodeString(msg.PublicKey)
	if err != nil {
		return nil, err
	}
	proof, err := hex.DecodeString(msg.Proof)
	if err != nil {
		return nil, err
	}
	_, err = vrf.NewK256VRF().Verify(publicKey, proof, hash)
	if err != nil {
		return nil, types.ErrInvalidStakeProof.Wrapf(err.Error())
	}

	// Verify that the staker is allowlisted if allowlist is enabled.
	params, err := m.Params.Get(ctx)
	if err != nil {
		return nil, err
	}
	if params.StakingConfig.AllowlistEnabled {
		allowlisted, err := m.Allowlist.Has(ctx, msg.PublicKey)
		if err != nil {
			return nil, err
		}
		if !allowlisted {
			return nil, types.ErrNotAllowlisted
		}
	}

	minStake := params.StakingConfig.MinimumStake
	denom, err := m.stakingKeeper.BondDenom(ctx)
	if err != nil {
		return nil, err
	}
	if msg.Stake.Denom != denom {
		return nil, sdkerrors.ErrInvalidCoins.Wrapf("invalid denom: %s", msg.Stake.Denom)
	}

	// Check stake amount and save the staker.
	if isExistingStaker {
		staker.Staked = staker.Staked.Add(msg.Stake)
		staker.Memo = msg.Memo
	} else {
		if msg.Stake.Amount.LT(minStake) {
			return nil, types.ErrInsufficientStake.Wrapf("%s is less than minimum stake %s", msg.Stake.Amount, minStake)
		}
		staker = types.Staker{
			PublicKey:         msg.PublicKey,
			Memo:              msg.Memo,
			Staked:            msg.Stake,
			PendingWithdrawal: sdk.NewInt64Coin(denom, 0),
			SequenceNum:       sequenceNum,
		}
	}

	senderAddr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}
	err = m.bankKeeper.SendCoinsFromAccountToModule(ctx, senderAddr, types.ModuleName, sdk.NewCoins(msg.Stake))
	if err != nil {
		return nil, err
	}

	staker.SequenceNum = sequenceNum + 1
	err = m.Stakers.Set(ctx, msg.PublicKey, staker)
	if err != nil {
		return nil, err
	}

	// TODO Add events

	return &types.MsgStakeResponse{}, nil
}

func (m msgServer) UpdateParams(goCtx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid authority address: %s", msg.Authority)
	}
	if m.GetAuthority() != msg.Authority {
		return nil, sdkerrors.ErrUnauthorized.Wrapf("unauthorized authority; expected %s, got %s", m.GetAuthority(), msg.Authority)
	}

	if err := msg.Params.Validate(); err != nil {
		return nil, err
	}
	if err := m.Params.Set(ctx, msg.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}
