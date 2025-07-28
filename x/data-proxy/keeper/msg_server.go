package keeper

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"

	"cosmossdk.io/collections"
	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/sedaprotocol/seda-chain/x/data-proxy/types"
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

func (m msgServer) RegisterDataProxy(goCtx context.Context, msg *types.MsgRegisterDataProxy) (*types.MsgRegisterDataProxyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	adminAddr, err := sdk.AccAddressFromBech32(msg.AdminAddress)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid admin address: %s", msg.AdminAddress)
	}

	// Burn the registration fee.
	params, err := m.GetParams(ctx)
	if err != nil {
		return nil, err
	}
	fee := sdk.NewCoins(params.RegistrationFee)

	err = m.bankKeeper.SendCoinsFromAccountToModule(ctx, adminAddr, types.ModuleName, fee)
	if err != nil {
		return nil, err
	}
	err = m.bankKeeper.BurnCoins(ctx, types.ModuleName, fee)
	if err != nil {
		return nil, err
	}

	pubKeyBytes, err := hex.DecodeString(msg.PubKey)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "invalid hex in pubkey: %s", msg.PubKey)
	}

	signatureBytes, err := hex.DecodeString(msg.Signature)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "invalid hex in signature: %s", msg.Signature)
	}

	found, err := m.HasDataProxy(ctx, pubKeyBytes)
	if err != nil {
		return nil, err
	}
	if found {
		return nil, types.ErrAlreadyExists
	}

	feeBytes := []byte(msg.Fee.String())
	adminAddressBytes := []byte(msg.AdminAddress)
	payoutAddressBytes := []byte(msg.PayoutAddress)
	memoBytes := []byte(msg.Memo)
	chainIDBytes := []byte(ctx.ChainID())

	payloadSize := len(feeBytes) + len(adminAddressBytes) + len(payoutAddressBytes) + len(memoBytes) + len(chainIDBytes)

	payload := make([]byte, 0, payloadSize)

	payload = append(payload, feeBytes...)
	payload = append(payload, adminAddressBytes...)
	payload = append(payload, payoutAddressBytes...)
	payload = append(payload, memoBytes...)
	payload = append(payload, chainIDBytes...)

	if valid := secp256k1.VerifySignature(pubKeyBytes, crypto.Keccak256(payload), signatureBytes); !valid {
		return nil, types.ErrInvalidSignature.Wrap("Invalid data proxy registration signature")
	}

	proxyConfig := types.ProxyConfig{
		PayoutAddress: msg.PayoutAddress,
		Fee:           msg.Fee,
		Memo:          msg.Memo,
		FeeUpdate:     nil,
		AdminAddress:  msg.AdminAddress,
	}

	err = proxyConfig.Validate()
	if err != nil {
		return nil, err
	}

	err = m.SetDataProxyConfig(ctx, pubKeyBytes, proxyConfig)
	if err != nil {
		return nil, err
	}

	emitProxyConfigEvent(ctx, types.EventTypeRegisterProxy, msg.PubKey, &proxyConfig)

	return &types.MsgRegisterDataProxyResponse{}, nil
}

func (m msgServer) EditDataProxy(goCtx context.Context, msg *types.MsgEditDataProxy) (*types.MsgEditDataProxyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	pubKeyBytes, err := hex.DecodeString(msg.PubKey)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "invalid hex in pubkey: %s", msg.PubKey)
	}

	proxyConfig, err := m.GetDataProxyConfig(ctx, pubKeyBytes)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, sdkerrors.ErrNotFound.Wrapf("no data proxy registered for %s", msg.PubKey)
		}
		return nil, err
	}

	if msg.Sender != proxyConfig.AdminAddress {
		return nil, sdkerrors.ErrorInvalidSigner
	}

	err = proxyConfig.UpdateBasic(msg.NewPayoutAddress, msg.NewMemo)
	if err != nil {
		return nil, err
	}

	// If there is no new fee we can terminate early
	if msg.NewFee == nil {
		err = m.SetDataProxyConfig(ctx, pubKeyBytes, proxyConfig)
		if err != nil {
			return nil, err
		}

		emitProxyConfigEvent(ctx, types.EventTypeEditProxy, msg.PubKey, &proxyConfig)

		return &types.MsgEditDataProxyResponse{}, nil
	}

	minimumUpdateDelay, err := m.GetMinimumUpdateDelay(ctx)
	if err != nil {
		return nil, err
	}

	updateDelay := minimumUpdateDelay
	// Validate custom delay if passed
	if msg.FeeUpdateDelay != types.UseMinimumDelay {
		if msg.FeeUpdateDelay < minimumUpdateDelay {
			return nil, types.ErrInvalidDelay.Wrapf("minimum delay %d, got %d", minimumUpdateDelay, msg.FeeUpdateDelay)
		}

		updateDelay = msg.FeeUpdateDelay
	}

	updateHeight, err := m.scheduleFeeUpdate(ctx, pubKeyBytes, proxyConfig, msg.NewFee, updateDelay)
	if err != nil {
		return nil, err
	}

	emitProxyConfigEvent(ctx, types.EventTypeEditProxy, msg.PubKey, &proxyConfig)

	return &types.MsgEditDataProxyResponse{
		FeeUpdateHeight: updateHeight,
	}, nil
}

func emitProxyConfigEvent(ctx sdk.Context, eventType string, pubKey string, proxyConfig *types.ProxyConfig) {
	event := sdk.NewEvent(eventType,
		sdk.NewAttribute(types.AttributePubKey, pubKey),
		sdk.NewAttribute(types.AttributePayoutAddress, proxyConfig.PayoutAddress),
		sdk.NewAttribute(types.AttributeFee, proxyConfig.Fee.String()),
		sdk.NewAttribute(types.AttributeMemo, proxyConfig.Memo),
		sdk.NewAttribute(types.AttributeAdminAddress, proxyConfig.AdminAddress),
	)

	if proxyConfig.FeeUpdate != nil {
		event.AppendAttributes(
			sdk.NewAttribute(types.AttributeNewFee, proxyConfig.FeeUpdate.String()),
			sdk.NewAttribute(types.AttributeNewFeeHeight, fmt.Sprintf("%d", proxyConfig.FeeUpdate.UpdateHeight)),
		)
	}

	ctx.EventManager().EmitEvent(event)
}

func (m msgServer) TransferAdmin(goCtx context.Context, msg *types.MsgTransferAdmin) (*types.MsgTransferAdminResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := msg.Validate(); err != nil {
		return nil, err
	}

	pubKeyBytes, err := hex.DecodeString(msg.PubKey)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "invalid hex in pubkey: %s", msg.PubKey)
	}

	proxyConfig, err := m.GetDataProxyConfig(ctx, pubKeyBytes)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, sdkerrors.ErrNotFound.Wrapf("no data proxy registered for %s", msg.PubKey)
		}
		return nil, err
	}

	if msg.Sender != proxyConfig.AdminAddress {
		return nil, sdkerrors.ErrorInvalidSigner
	}

	proxyConfig.AdminAddress = msg.NewAdminAddress

	err = m.SetDataProxyConfig(ctx, pubKeyBytes, proxyConfig)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(types.EventTypeTransferAdmin,
		sdk.NewAttribute(types.AttributePubKey, msg.PubKey),
		sdk.NewAttribute(types.AttributeAdminAddress, proxyConfig.AdminAddress),
	))

	return &types.MsgTransferAdminResponse{}, nil
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
	if err := m.SetParams(ctx, msg.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}
