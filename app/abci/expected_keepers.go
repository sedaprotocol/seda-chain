package abci

import (
	"context"

	"github.com/cosmos/cosmos-sdk/baseapp"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/sedaprotocol/seda-chain/app/utils"
	"github.com/sedaprotocol/seda-chain/x/batching/types"
)

type BatchingKeeper interface {
	GetBatchForHeight(ctx context.Context, height int64) (types.Batch, error)
	SetBatchSignatures(ctx context.Context, batchNum uint64, sigs types.BatchSignatures) error
}

type PubKeyKeeper interface {
	GetValidatorKeyAtIndex(ctx context.Context, valAddr sdk.ValAddress, index utils.SEDAKeyIndex) (cryptotypes.PubKey, error)
}

type StakingKeeper interface {
	baseapp.ValidatorStore
	GetValidatorByConsAddr(ctx context.Context, consAddr sdk.ConsAddress) (stakingtypes.Validator, error)
}
