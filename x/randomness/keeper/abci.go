package keeper

import (
	"encoding/hex"
	"strings"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/x/randomness/types"
)

func (k *Keeper) EndBlocker(ctx sdk.Context) []abci.ValidatorUpdate {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyEndBlocker)
	k.Logger(ctx).Info("adding app hash %s", string(ctx.BlockHeader().AppHash))
	k.SetSeed(ctx, strings.ToUpper(hex.EncodeToString(ctx.BlockHeader().AppHash)))
	return nil
}
