package keeper_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	keepertest "github.com/sedaprotocol/seda-chain/testutil/keeper"
	"github.com/sedaprotocol/seda-chain/x/sedachain/keeper"
	"github.com/sedaprotocol/seda-chain/x/sedachain/types"
)

func setupMsgServer(t *testing.T) (types.MsgServer, context.Context) {
	t.Helper()

	k, ctx := keepertest.SedachainKeeper(t)
	return keeper.NewMsgServerImpl(*k), sdk.WrapSDKContext(ctx)
}

func TestMsgServer(t *testing.T) {
	ms, ctx := setupMsgServer(t)
	require.NotNil(t, ms)
	require.NotNil(t, ctx)
}
