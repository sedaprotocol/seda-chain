package types

import (
	"context"

	"cosmossdk.io/core/address"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BankKeeper defines the contract needed for supply related APIs (noalias)
type BankKeeper interface {
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	BurnCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
}

type AccountKeeper interface {
	AddressCodec() address.Codec

	GetModuleAddress(name string) sdk.AccAddress
	GetModuleAccount(ctx context.Context, moduleName string) sdk.ModuleAccountI
}

type StakingKeeper interface {
	BondDenom(ctx context.Context) (string, error)
}
