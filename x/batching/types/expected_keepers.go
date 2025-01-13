package types

import (
	"context"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/sedaprotocol/seda-chain/app/utils"
)

type SlashingKeeper interface {
	IsTombstoned(ctx context.Context, consAddr sdk.ConsAddress) bool
	SlashFractionDoubleSign(ctx context.Context) (math.LegacyDec, error)
	JailUntil(ctx context.Context, consAddr sdk.ConsAddress, jailTime time.Time) error
	Jail(ctx context.Context, consAddr sdk.ConsAddress) error
	Tombstone(ctx context.Context, consAddr sdk.ConsAddress) error
}

type StakingKeeper interface {
	GetBondedValidatorsByPower(ctx context.Context) ([]stakingtypes.Validator, error)
	GetValidatorUpdates(ctx context.Context) ([]abci.ValidatorUpdate, error)
	IterateLastValidatorPowers(ctx context.Context, handler func(operator sdk.ValAddress, power int64) (stop bool)) error
	GetLastTotalPower(ctx context.Context) (math.Int, error)
	GetHistoricalInfo(ctx context.Context, height int64) (stakingtypes.HistoricalInfo, error)
	Slash(ctx context.Context, consAddr sdk.ConsAddress, infractionHeight, power int64, slashFactor math.LegacyDec) (math.Int, error)
}

type WasmStorageKeeper interface {
	GetCoreContractAddr(ctx context.Context) (sdk.AccAddress, error)
}

type PubKeyKeeper interface {
	GetValidatorKeyAtIndex(ctx context.Context, validatorAddr sdk.ValAddress, index utils.SEDAKeyIndex) ([]byte, error)
	IsProvingSchemeActivated(ctx context.Context, index utils.SEDAKeyIndex) (bool, error)
}
