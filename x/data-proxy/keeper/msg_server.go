package keeper

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"

	sdk "github.com/cosmos/cosmos-sdk/types"

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

	if err := msg.Validate(); err != nil {
		return nil, err
	}

	if _, err := sdk.AccAddressFromBech32(msg.AdminAddress); err != nil {
		return nil, types.ErrInvalidAddress.Wrapf("invalid admin address %s", err)
	}

	pubKeyBytes, err := hex.DecodeString(msg.PubKey)
	if err != nil {
		return nil, types.ErrInvalidHex.Wrap(err.Error())
	}

	signatureBytes, err := hex.DecodeString(msg.Signature)
	if err != nil {
		return nil, types.ErrInvalidHex.Wrap(err.Error())
	}

	found, err := m.DataProxyConfigs.Has(ctx, pubKeyBytes)
	if err != nil {
		return nil, err
	}
	if found {
		return nil, types.ErrAlreadyExists
	}

	feeBytes := []byte(msg.Fee.String())
	payoutAddressBytes := []byte(msg.PayoutAddress)
	memoBytes := []byte(msg.Memo)

	payload := make([]byte, 0, len(feeBytes)+len(payoutAddressBytes)+len(memoBytes))
	payload = append(payload, feeBytes...)
	payload = append(payload, payoutAddressBytes...)
	payload = append(payload, memoBytes...)

	if valid := secp256k1.VerifySignature(pubKeyBytes, crypto.Keccak256(payload), signatureBytes); !valid {
		return nil, types.ErrInvalidSignature
	}

	err = m.DataProxyConfigs.Set(ctx, pubKeyBytes, types.ProxyConfig{
		PayoutAddress: msg.PayoutAddress,
		Fee:           msg.Fee,
		Memo:          msg.Memo,
		FeeUpdate:     nil,
		AdminAddress:  msg.AdminAddress,
	})
	if err != nil {
		return nil, err
	}

	return &types.MsgRegisterDataProxyResponse{}, nil
}

func (m msgServer) EditDataProxy(_ context.Context, _ *types.MsgEditDataProxy) (*types.MsgEditDataProxyResponse, error) {
	// TODO
	return &types.MsgEditDataProxyResponse{}, nil
}

func (m msgServer) UpdateParams(goCtx context.Context, req *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if _, err := sdk.AccAddressFromBech32(req.Authority); err != nil {
		return nil, fmt.Errorf("invalid authority address: %s", err)
	}
	if m.GetAuthority() != req.Authority {
		return nil, fmt.Errorf("unauthorized authority; expected %s, got %s", m.GetAuthority(), req.Authority)
	}

	if err := req.Params.Validate(); err != nil {
		return nil, err
	}
	if err := m.Params.Set(ctx, req.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}
