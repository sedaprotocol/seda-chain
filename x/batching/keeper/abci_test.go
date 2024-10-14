package keeper_test

import (
	cryptorand "crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/sha3"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkstakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	sdkstakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/sedaprotocol/seda-chain/app/utils"
	"github.com/sedaprotocol/seda-chain/x/batching/types"
)

func Test_ConstructDataResultTree(t *testing.T) {
	f := initFixture(t)

	dataResults := generateDataResults(t, 25)
	for _, dr := range dataResults {
		err := f.batchingKeeper.SetDataResultForBatching(f.Context(), dr)
		require.NoError(t, err)
	}

	entries, root, err := f.batchingKeeper.ConstructDataResultTree(f.Context(), rand.Uint64())
	require.NoError(t, err)

	var entryHexes, drIds []string
	for i, entry := range entries {
		entryHexes = append(entryHexes, hex.EncodeToString(entry))
		drIds = append(drIds, dataResults[i].Id)
	}
	require.ElementsMatch(t, drIds, entryHexes)

	// Generate proof for each entry and verify.
	require.NoError(t, err)
	for i := range entries {
		proof, err := utils.GetProof(entries, i)
		require.NoError(t, err)

		// TODO: Generalize and provide a utils function
		sha := sha3.NewLegacyKeccak256()
		sha.Write([]byte{})
		emptyLeaf := sha.Sum(nil)
		proof = append(proof, emptyLeaf)

		ret := utils.VerifyProof(proof, root, entries[i])
		require.True(t, ret)
	}
}

func Test_ConstructValidatorTree(t *testing.T) {
	f := initFixture(t)
	_, pks, powers := addBatchSigningValidators(t, f, 10)

	entries, root, err := f.batchingKeeper.ConstructValidatorTree(f.Context())
	require.NoError(t, err)

	var totalPower int64
	for _, power := range powers {
		totalPower += power
	}
	var powerPercents []uint32
	for i := range powers {
		powerPercents = append(powerPercents, uint32(powers[i]*1e8/totalPower))
	}

	// Parse entries to check contents.
	parsedPKs := make([][]byte, len(entries))
	parsedPowers := make([]uint32, len(entries))
	for i, entry := range entries {
		require.Equal(t, []byte("SECP256K1"), entry[:9])
		parsedPKs[i] = entry[9:41]
		parsedPowers[i] = binary.BigEndian.Uint32(entry[41:])
	}
	require.ElementsMatch(t, pks, parsedPKs)
	require.ElementsMatch(t, powerPercents, parsedPowers)

	// Generate proof for each entry and verify.
	require.NoError(t, err)
	for i := range entries {
		pf, err := utils.GetProof(entries, i)
		require.NoError(t, err)
		ret := utils.VerifyProof(pf, root, entries[i])
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

// generateDataResults returns a given number of randomly-generated
// data results.
func generateDataResults(t *testing.T, num int) []types.DataResult {
	t.Helper()
	dataResults := make([]types.DataResult, num)
	for i := range dataResults {
		sample := types.DataResult{
			DrId:           generateRandomHexString(64),
			Version:        fmt.Sprintf("%d.%d.%d", rand.Intn(10), rand.Intn(10), rand.Intn(10)),
			BlockHeight:    rand.Uint64(),
			ExitCode:       rand.Uint32(),
			GasUsed:        rand.Uint64(),
			Result:         generateRandomBytes(50),
			PaybackAddress: generateRandomBase64String(10),
			SedaPayload:    generateRandomBase64String(10),
			Consensus:      rand.Intn(2) == 1,
		}

		var err error
		sample.Id, err = sample.TryHash()
		require.NoError(t, err)

		dataResults[i] = sample
	}
	return dataResults
}

func generateRandomHexString(length int) string {
	bytes := make([]byte, length/2)
	cryptorand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func generateRandomBase64String(length int) string {
	bytes := make([]byte, length)
	cryptorand.Read(bytes)
	return base64.StdEncoding.EncodeToString(bytes)
}

func generateRandomBytes(length int) []byte {
	bytes := make([]byte, length)
	cryptorand.Read(bytes)
	return bytes
}
