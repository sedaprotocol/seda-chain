package abci

import (
	"context"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	sedatypes "github.com/sedaprotocol/seda-chain/types"
	batchingtypes "github.com/sedaprotocol/seda-chain/x/batching/types"
	pubkeytypes "github.com/sedaprotocol/seda-chain/x/pubkey/types"
)

type BatchingKeeper interface {
	GetBatchForHeight(ctx context.Context, height int64) (batchingtypes.Batch, error)
	SetBatchSigSecp256k1(ctx context.Context, batchNum uint64, valAddr sdk.ValAddress, signature []byte) error
	GetValidatorTreeEntry(ctx context.Context, batchNum uint64, valAddr sdk.ValAddress) (batchingtypes.ValidatorTreeEntry, error)
}

type PubKeyKeeper interface {
	GetValidatorKeys(ctx context.Context, validatorAddr string) (result pubkeytypes.ValidatorPubKeys, err error)
	GetValidatorKeyAtIndex(ctx context.Context, validatorAddr sdk.ValAddress, index sedatypes.SEDAKeyIndex) ([]byte, error)
}

type StakingKeeper interface {
	baseapp.ValidatorStore
	GetValidator(ctx context.Context, addr sdk.ValAddress) (stakingtypes.Validator, error)
	GetValidatorByConsAddr(ctx context.Context, consAddr sdk.ConsAddress) (stakingtypes.Validator, error)
}
