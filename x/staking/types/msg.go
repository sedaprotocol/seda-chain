package types

import (
	"cosmossdk.io/core/address"
	"cosmossdk.io/math"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/staking/types"

	pubkeytypes "github.com/sedaprotocol/seda-chain/x/pubkey/types"
)

var (
	_ sdk.Msg                            = &MsgCreateSEDAValidator{}
	_ codectypes.UnpackInterfacesMessage = (*MsgCreateSEDAValidator)(nil)
)

// NewMsgCreateSEDAValidator creates a MsgCreateSEDAValidator instance.
func NewMsgCreateSEDAValidator(
	valAddr string, pubKey cryptotypes.PubKey, sedaPubKeys []pubkeytypes.IndexedPubKey,
	selfDelegation sdk.Coin, description types.Description, commission types.CommissionRates,
	minSelfDelegation math.Int,
) (*MsgCreateSEDAValidator, error) {
	var pkAny *codectypes.Any
	if pubKey != nil {
		var err error
		if pkAny, err = codectypes.NewAnyWithValue(pubKey); err != nil {
			return nil, err
		}
	}

	return &MsgCreateSEDAValidator{
		Description:       description,
		ValidatorAddress:  valAddr,
		Pubkey:            pkAny,
		IndexedPubKeys:    sedaPubKeys,
		Value:             selfDelegation,
		Commission:        commission,
		MinSelfDelegation: minSelfDelegation,
	}, nil
}

// Validate validates the MsgCreateSEDAValidator sdk msg.
func (msg MsgCreateSEDAValidator) Validate(ac address.Codec) error {
	if msg.Pubkey == nil {
		return sdkerrors.ErrInvalidRequest.Wrap("empty validator public key")
	}

	msgCreateVal := &types.MsgCreateValidator{
		Description:       msg.Description,
		Commission:        msg.Commission,
		MinSelfDelegation: msg.MinSelfDelegation,
		ValidatorAddress:  msg.ValidatorAddress,
		Pubkey:            msg.Pubkey,
		Value:             msg.Value,
	}
	return msgCreateVal.Validate(ac)
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (msg MsgCreateSEDAValidator) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	var pubKey cryptotypes.PubKey
	return unpacker.UnpackAny(msg.Pubkey, &pubKey)
}
