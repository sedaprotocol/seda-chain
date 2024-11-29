package types

import (
	"context"

	abci "github.com/cometbft/cometbft/abci/types"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/sedaprotocol/seda-chain/app/utils"
)

type StakingKeeper interface {
	GetBondedValidatorsByPower(ctx context.Context) ([]stakingtypes.Validator, error)
	GetValidatorUpdates(ctx context.Context) ([]abci.ValidatorUpdate, error)
	IterateLastValidatorPowers(ctx context.Context, handler func(operator sdk.ValAddress, power int64) (stop bool)) error
	GetLastTotalPower(ctx context.Context) (math.Int, error)
}

type WasmStorageKeeper interface {
	GetCoreContractAddr(ctx context.Context) (sdk.AccAddress, error)
}

type PubKeyKeeper interface {
	GetValidatorKeyAtIndex(ctx context.Context, validatorAddr sdk.ValAddress, index utils.SEDAKeyIndex) ([]byte, error)
}
