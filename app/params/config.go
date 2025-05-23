package params

import (
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	HumanCoinUnit    = "seda"
	BaseCoinUnit     = "aseda" // atto (10^-18)
	DefaultBondDenom = BaseCoinUnit
	SedaExponent     = 18

	// Bech32PrefixAccAddr defines the Bech32 prefix of an account's address.
	Bech32PrefixAccAddr = "seda"

	DefaultMempoolMaxTxs = 5000
)

var (
	// Bech32PrefixAccPub defines the Bech32 prefix of an account's public key.
	Bech32PrefixAccPub = Bech32PrefixAccAddr + "pub"
	// Bech32PrefixValAddr defines the Bech32 prefix of a validator's operator address.
	Bech32PrefixValAddr = Bech32PrefixAccAddr + "valoper"
	// Bech32PrefixValPub defines the Bech32 prefix of a validator's operator public key.
	Bech32PrefixValPub = Bech32PrefixAccAddr + "valoperpub"
	// Bech32PrefixConsAddr defines the Bech32 prefix of a consensus node address.
	Bech32PrefixConsAddr = Bech32PrefixAccAddr + "valcons"
	// Bech32PrefixConsPub defines the Bech32 prefix of a consensus node public key.
	Bech32PrefixConsPub = Bech32PrefixAccAddr + "valconspub"

	MinimumGasPrice = sdk.NewDecCoinFromDec(BaseCoinUnit, math.LegacyMustNewDecFromStr("10000000000")) // 10^10 aseda
)

func init() {
	SetAddressPrefixes()
	RegisterDenoms()
}

// RegisterDenoms registers token denoms.
func RegisterDenoms() {
	err := sdk.RegisterDenom(BaseCoinUnit, math.LegacyNewDecWithPrec(1, SedaExponent))
	if err != nil {
		panic(err)
	}

	err = sdk.RegisterDenom(HumanCoinUnit, math.LegacyOneDec())
	if err != nil {
		panic(err)
	}
}

// SetAddressPrefixes builds the Config with Bech32 addressPrefix and publKeyPrefix for accounts, validators, and consensus nodes and verifies that addreeses have correct format.
func SetAddressPrefixes() {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(Bech32PrefixAccAddr, Bech32PrefixAccPub)
	config.SetBech32PrefixForValidator(Bech32PrefixValAddr, Bech32PrefixValPub)
	config.SetBech32PrefixForConsensusNode(Bech32PrefixConsAddr, Bech32PrefixConsPub)

	// This is copied from the cosmos sdk v0.43.0-beta1
	// source: https://github.com/cosmos/cosmos-sdk/blob/v0.43.0-beta1/types/address.go#L141
	config.SetAddressVerifier(func(bytes []byte) error {
		if len(bytes) == 0 {
			return errorsmod.Wrap(sdkerrors.ErrUnknownAddress, "addresses cannot be empty")
		}

		if len(bytes) > address.MaxAddrLen {
			return errorsmod.Wrapf(sdkerrors.ErrUnknownAddress, "address max length is %d, got %d, %x", address.MaxAddrLen, len(bytes), bytes)
		}

		if len(bytes) != 20 && len(bytes) != 32 {
			return errorsmod.Wrapf(sdkerrors.ErrUnknownAddress, "address length must be 20 or 32 bytes, got %d, %x", len(bytes), bytes)
		}

		return nil
	})
}
