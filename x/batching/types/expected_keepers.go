package types

import (
	"context"

	abci "github.com/cometbft/cometbft/abci/types"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

type StakingKeeper interface {
	GetBondedValidatorsByPower(ctx context.Context) ([]stakingtypes.Validator, error)
	GetValidatorUpdates(ctx context.Context) ([]abci.ValidatorUpdate, error)
	IterateLastValidatorPowers(ctx context.Context, handler func(operator sdk.ValAddress, power int64) (stop bool)) error
}

type WasmStorageKeeper interface {
	GetCoreContractAddr(ctx context.Context) (sdk.AccAddress, error)
}

type PubKeyKeeper interface {
	GetValidatorKeyAtIndex(ctx context.Context, validatorAddr sdk.ValAddress, index uint32) (cryptotypes.PubKey, error)
}
