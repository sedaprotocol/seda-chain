package types

import (
	context "context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	sedatypes "github.com/sedaprotocol/seda-chain/types"
	"github.com/sedaprotocol/seda-chain/x/pubkey/types"
)

type PubKeyKeeper interface {
	StoreIndexedPubKeys(ctx sdk.Context, valAddr sdk.ValAddress, pubKeys []types.IndexedPubKey) error
	IsProvingSchemeActivated(ctx context.Context, index sedatypes.SEDAKeyIndex) (bool, error)
	HasRegisteredKey(ctx context.Context, validatorAddr sdk.ValAddress, index sedatypes.SEDAKeyIndex) (bool, error)
}
