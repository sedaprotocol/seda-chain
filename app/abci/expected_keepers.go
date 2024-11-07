package abci

import (
	"context"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/sedaprotocol/seda-chain/app/utils"
	batchingtypes "github.com/sedaprotocol/seda-chain/x/batching/types"
	pubkeytypes "github.com/sedaprotocol/seda-chain/x/pubkey/types"
)

type BatchingKeeper interface {
	GetBatchForHeight(ctx context.Context, height int64) (batchingtypes.Batch, error)
	SetBatchSignatures(ctx context.Context, sigs batchingtypes.BatchSignatures) error
	GetValidatorTreeEntry(ctx context.Context, batchNum uint64, valAddress sdk.ValAddress) ([]byte, error)
}

type PubKeyKeeper interface {
	GetValidatorKeys(ctx context.Context, validatorAddr string) (result pubkeytypes.ValidatorPubKeys, err error)
	GetValidatorKeyAtIndex(ctx context.Context, validatorAddr sdk.ValAddress, index utils.SEDAKeyIndex) ([]byte, error)
}

type StakingKeeper interface {
	baseapp.ValidatorStore
	GetValidatorByConsAddr(ctx context.Context, consAddr sdk.ConsAddress) (stakingtypes.Validator, error)
}
