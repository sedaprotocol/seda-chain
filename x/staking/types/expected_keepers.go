package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/sedaprotocol/seda-chain/x/pubkey/types"
)

type PubkeyKeeper interface {
	StoreIndexedPubKeys(ctx sdk.Context, valAddr sdk.ValAddress, pubKeys []types.IndexedPubKey) error
}
