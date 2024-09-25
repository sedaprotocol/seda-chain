package keeper_test

import (
	"encoding/binary"
	"encoding/hex"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkstakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	sdkstakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/sedaprotocol/seda-chain/cmd/sedad/utils"
)

func Test_ConstructValidatorTree(t *testing.T) {
	f := initFixture(t)
	_, pks, powers := addBatchSigningValidators(t, f, 100)

	entries, root, err := f.batchingKeeper.ConstructValidatorTree(f.Context())
	require.NoError(t, err)

	// Parse entries to check contents.
	parsedPKs := make([][]byte, len(entries))
	parsedPowers := make([]int64, len(entries))
	for i, entry := range entries {
		parsedPKs[i] = entry[:32]
		parsedPowers[i] = int64(binary.BigEndian.Uint64(entry[32:]))
	}
	require.ElementsMatch(t, pks, parsedPKs)
	require.ElementsMatch(t, powers, parsedPowers)

	// Generate proof for each entry and verify.
	rootBytes, err := hex.DecodeString(root)
	require.NoError(t, err)
	for i := range entries {
		pf, err := utils.GetProof(entries, i)
		require.NoError(t, err)
		ret := utils.VerifyProof(pf, rootBytes, entries[i])
		require.True(t, ret)
	}
}

// addBatchSigningValidators funds test addresses, adds them as validators,
// and registers secp256k1 public keys for their batch signing.
func addBatchSigningValidators(t *testing.T, f *fixture, num int) ([]sdk.AccAddress, [][]byte, []int64) {
	t.Helper()

	ctx := f.Context()
	stakingMsgSvr := sdkstakingkeeper.NewMsgServerImpl(f.stakingKeeper.Keeper)

	addrs := simtestutil.AddTestAddrs(f.bankKeeper, f.stakingKeeper, ctx, num, math.NewIntFromUint64(10000000000000000000))
	pubKeys := make([][]byte, len(addrs))
	powers := make([]int64, len(addrs))
	for i, addr := range addrs {
		valAddr := sdk.ValAddress(addr)
		pubKey := ed25519.GenPrivKey().PubKey()
		pubKeys[i] = pubKey.Bytes()
		powers[i] = int64(rand.Intn(10) + 1)

		valTokens := sdk.TokensFromConsensusPower(powers[i], sdk.DefaultPowerReduction)
		valCreateMsg, err := sdkstakingtypes.NewMsgCreateValidator(
			valAddr.String(),
			pubKey,
			sdk.NewCoin(bondDenom, valTokens),
			sdkstakingtypes.NewDescription("T", "E", "S", "T", "Z"),
			sdkstakingtypes.NewCommissionRates(math.LegacyZeroDec(), math.LegacyZeroDec(), math.LegacyZeroDec()),
			math.OneInt(),
		)
		require.NoError(t, err)
		res, err := stakingMsgSvr.CreateValidator(ctx, valCreateMsg)
		require.NoError(t, err)
		require.NotNil(t, res)

		_, err = f.stakingKeeper.Keeper.EndBlocker(ctx)
		require.NoError(t, err)

		err = f.pubKeyKeeper.SetValidatorKeyAtIndex(ctx, valAddr, 0, pubKey)
		require.NoError(t, err)
	}
	return addrs, pubKeys, powers
}
