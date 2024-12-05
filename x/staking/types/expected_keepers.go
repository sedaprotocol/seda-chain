package types

import (
	context "context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/app/utils"
	"github.com/sedaprotocol/seda-chain/x/pubkey/types"
)

type PubKeyKeeper interface {
	StoreIndexedPubKeys(ctx sdk.Context, valAddr sdk.ValAddress, pubKeys []types.IndexedPubKey) error
	IsProvingSchemeActivated(ctx context.Context, index utils.SEDAKeyIndex) (bool, error)
	GetValidatorKeyAtIndex(ctx context.Context, validatorAddr sdk.ValAddress, index utils.SEDAKeyIndex) ([]byte, error)
}
