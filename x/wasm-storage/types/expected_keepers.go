package types

import (
	context "context"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type BankKeeper interface {
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
}

type StakingKeeper interface {
	BondDenom(ctx context.Context) (string, error)
}
