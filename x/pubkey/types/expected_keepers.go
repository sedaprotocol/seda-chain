package types

import (
	"context"

	abci "github.com/cometbft/cometbft/abci/types"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

type StakingKeeper interface {
	GetValidator(ctx context.Context, addr sdk.ValAddress) (validator stakingtypes.Validator, err error)
	GetBondedValidatorsByPower(ctx context.Context) ([]stakingtypes.Validator, error)
	GetValidatorUpdates(ctx context.Context) ([]abci.ValidatorUpdate, error)
	IterateLastValidatorPowers(ctx context.Context, handler func(operator sdk.ValAddress, power int64) (stop bool)) error
	GetLastTotalPower(ctx context.Context) (math.Int, error)
}
