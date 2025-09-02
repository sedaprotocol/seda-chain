package types

import (
	"context"

	"cosmossdk.io/core/address"

	sdk "github.com/cosmos/cosmos-sdk/types"

	dataproxytypes "github.com/sedaprotocol/seda-chain/x/data-proxy/types"
)

// BankKeeper defines the contract needed for supply related APIs (noalias)
type BankKeeper interface {
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
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

type DataProxyKeeper interface {
	GetDataProxyConfig(ctx context.Context, pubKey []byte) (result dataproxytypes.ProxyConfig, err error)
}
