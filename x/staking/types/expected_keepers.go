package types

import (
	"context"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

type AccountKeeper interface {
	stakingtypes.AccountKeeper

	NewAccountWithAddress(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
	SetAccount(ctx context.Context, acc sdk.AccountI)
}

type RandomnessKeeper interface {
	SetValidatorVRFPubKey(ctx context.Context, consensusAddr string, vrfPubKey cryptotypes.PubKey) error
}
