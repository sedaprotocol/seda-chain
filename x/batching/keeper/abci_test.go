package keeper_test

import (
	cryptorand "crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"testing"

	dcrdsecp256k1 "github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/sha3"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
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
		entryHexes = append(entryHexes, hex.EncodeToString(entry[1:]))
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
	uncompressedPKs := make([][]byte, len(entries))
	parsedPKs := make([][]byte, len(entries))
	parsedPowers := make([]uint32, len(entries))
	for i, entry := range entries {
		require.Equal(t, []byte{utils.SEDASeparatorSecp256k1}, entry[:1])
		parsedPKs[i] = entry[1:66]
		parsedPowers[i] = binary.BigEndian.Uint32(entry[66:])

		uncompressedPKs[i] = decompressPubKey(t, pks[i])
	}
	require.ElementsMatch(t, uncompressedPKs, parsedPKs)
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
		pubKey := secp256k1.GenPrivKey().PubKey()
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

		err = f.pubKeyKeeper.SetValidatorKeyAtIndex(ctx, valAddr, utils.SEDAKeyIndexSecp256k1, pubKey)
		require.NoError(t, err)
	}
	return addrs, pubKeys, powers
}

// decompressPubKey decompresses a 33-byte long compressed public key
// into a 65-byte long uncompressed format.
func decompressPubKey(t *testing.T, pubKey []byte) []byte {
	t.Helper()
	pk, err := dcrdsecp256k1.ParsePubKey(pubKey)
	if err != nil {
		panic(err)
	}
	return pk.SerializeUncompressed()
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

type Result struct {
	ResultID       string `json:"resultId"`
	Version        string `json:"version"`
	DrID           string `json:"drId"`
	Consensus      bool   `json:"consensus"`
	ExitCode       int    `json:"exitCode"`
	Result         string `json:"result"`
	BlockHeight    int    `json:"blockHeight"`
	GasUsed        int    `json:"gasUsed"`
	PaybackAddress string `json:"paybackAddress"`
	SedaPayload    string `json:"sedaPayload"`
}

type Tree struct {
	Root   string   `json:"root"`
	Leaves []string `json:"leaves"`
}

type TestData struct {
	Results []Result `json:"results"`
	Tree    Tree     `json:"tree"`
}

const testDataJSON = `{
  "results": [
    {
      "resultId": "0xbe65b686b9bd9a896762f4c6fef234290ec55e3ab609426c685c53d61287d336",
      "version": "0.0.1",
      "drId": "0x2e469f40cf0afb5b93641c75aaecb654f0c73fc2dd4fcc350150c4825d2fb36e",
      "consensus": true,
      "exitCode": 0,
      "result": "0x39bf027dd97f3bae0cf8cfb909695ec63313a9bd61ad52fc7f52cf565b141da8",
      "blockHeight": 0,
      "gasUsed": 0,
      "paybackAddress": "0x0000000000000000000000000000000000000000",
      "sedaPayload": "0x0000000000000000000000000000000000000000000000000000000000000000"
    },
    {
      "resultId": "0x99de4dde0fb83fa812b9b8c60b2cf9d47518e456b83bd42eb1c7d98bbb7ede58",
      "version": "0.0.1",
      "drId": "0xa779f7061360e9fd128acb8686c393294db943848dc6c1e2248dd91fc79f3835",
      "consensus": true,
      "exitCode": 0,
      "result": "0x39bf027dd97f3bae0cf8cfb909695ec63313a9bd61ad52fc7f52cf565b141da8",
      "blockHeight": 0,
      "gasUsed": 0,
      "paybackAddress": "0x0000000000000000000000000000000000000000",
      "sedaPayload": "0x0000000000000000000000000000000000000000000000000000000000000000"
    },
    {
      "resultId": "0xcf9138035c2c57694000aa70218b2523eefe150257482b9af18d4633e7e8b97b",
      "version": "0.0.1",
      "drId": "0xcedae320cef23e6b50baf01f722b53818fd1d738d796a57a785c71436237dd35",
      "consensus": true,
      "exitCode": 0,
      "result": "0x39bf027dd97f3bae0cf8cfb909695ec63313a9bd61ad52fc7f52cf565b141da8",
      "blockHeight": 0,
      "gasUsed": 0,
      "paybackAddress": "0x0000000000000000000000000000000000000000",
      "sedaPayload": "0x0000000000000000000000000000000000000000000000000000000000000000"
    },
    {
      "resultId": "0xc0b86bcf24a3e305b0319e45106b7851ee8e59f7e2d9793a886c16a8949ca898",
      "version": "0.0.1",
      "drId": "0xd2e85a4f932ee01d3aec08c4fb0725425073914e11e32219a90d0fb190adcce3",
      "consensus": true,
      "exitCode": 0,
      "result": "0x39bf027dd97f3bae0cf8cfb909695ec63313a9bd61ad52fc7f52cf565b141da8",
      "blockHeight": 0,
      "gasUsed": 0,
      "paybackAddress": "0x0000000000000000000000000000000000000000",
      "sedaPayload": "0x0000000000000000000000000000000000000000000000000000000000000000"
    },
    {
      "resultId": "0x7d2731c2c215d7a90f8037e7ce7fb5dfb8699417fc3cda6a946fcc13761c8b77",
      "version": "0.0.1",
      "drId": "0xa11695dd090839580ee64a21412e8f0843255c149f5352311ab6960990e690be",
      "consensus": true,
      "exitCode": 0,
      "result": "0x39bf027dd97f3bae0cf8cfb909695ec63313a9bd61ad52fc7f52cf565b141da8",
      "blockHeight": 0,
      "gasUsed": 0,
      "paybackAddress": "0x0000000000000000000000000000000000000000",
      "sedaPayload": "0x0000000000000000000000000000000000000000000000000000000000000000"
    },
    {
      "resultId": "0xc6c8c01a175383783db75dd7199cbe74db15438880e5fe5e7e6dcbe9fbc96e4c",
      "version": "0.0.1",
      "drId": "0xa4a1baa9d463128792770c0aadfc7916362d9309e19d709729060a48eb3b4b13",
      "consensus": true,
      "exitCode": 0,
      "result": "0x39bf027dd97f3bae0cf8cfb909695ec63313a9bd61ad52fc7f52cf565b141da8",
      "blockHeight": 0,
      "gasUsed": 0,
      "paybackAddress": "0x0000000000000000000000000000000000000000",
      "sedaPayload": "0x0000000000000000000000000000000000000000000000000000000000000000"
    },
    {
      "resultId": "0x998c975ddc39a0185b02e544001f63eac401b01fe3737c4f9da4992bcb146bea",
      "version": "0.0.1",
      "drId": "0xe220988adad2c99912c5a4a7feb197e04a1a0396de7ca16b5e47dc1b6995b484",
      "consensus": true,
      "exitCode": 0,
      "result": "0x39bf027dd97f3bae0cf8cfb909695ec63313a9bd61ad52fc7f52cf565b141da8",
      "blockHeight": 0,
      "gasUsed": 0,
      "paybackAddress": "0x0000000000000000000000000000000000000000",
      "sedaPayload": "0x0000000000000000000000000000000000000000000000000000000000000000"
    },
    {
      "resultId": "0x57d66709949b25a61c29a914ca6b63680b95843f16c9e6f796ebb3ede2ca0f58",
      "version": "0.0.1",
      "drId": "0x2d9d50f861512a9f7041452eddfbff39f2553e1ebb5646d2eecf6733d7c8d2ce",
      "consensus": true,
      "exitCode": 0,
      "result": "0x39bf027dd97f3bae0cf8cfb909695ec63313a9bd61ad52fc7f52cf565b141da8",
      "blockHeight": 0,
      "gasUsed": 0,
      "paybackAddress": "0x0000000000000000000000000000000000000000",
      "sedaPayload": "0x0000000000000000000000000000000000000000000000000000000000000000"
    },
    {
      "resultId": "0xe3c54225c36fe4b1250fb806ef1cc5c13f291069941259a4c58664419c222528",
      "version": "0.0.1",
      "drId": "0x3b8dfa0893f0a24ce3d600fad4ca16fd08ee257efea866145d000e65cf14f04a",
      "consensus": true,
      "exitCode": 0,
      "result": "0x39bf027dd97f3bae0cf8cfb909695ec63313a9bd61ad52fc7f52cf565b141da8",
      "blockHeight": 0,
      "gasUsed": 0,
      "paybackAddress": "0x0000000000000000000000000000000000000000",
      "sedaPayload": "0x0000000000000000000000000000000000000000000000000000000000000000"
    },
    {
      "resultId": "0xafb0ab28855671a35c30b99eab2fd1e2a71486bfc6fc625638fd31a2ac5f26ed",
      "version": "0.0.1",
      "drId": "0xf180cde27633a009ad05fe86b0ef69c7f2065c3fcf40b103710173829990ce87",
      "consensus": true,
      "exitCode": 0,
      "result": "0x39bf027dd97f3bae0cf8cfb909695ec63313a9bd61ad52fc7f52cf565b141da8",
      "blockHeight": 0,
      "gasUsed": 0,
      "paybackAddress": "0x0000000000000000000000000000000000000000",
      "sedaPayload": "0x0000000000000000000000000000000000000000000000000000000000000000"
    }
  ],
  "tree": {
    "root": "0xf0fd6bcebfba525736179ce169ae7a219bf994857bc51546fb83661be3a61744",
    "leaves": [
      "0xae813aa6da400689d42228fb2f352734d0f66114bd1755446f1217456e22b4ad",
      "0x32047ba1e49f5435144d16d581c11851989a357cdf2924f53cb4322c5c92f680",
      "0xb5a860ffc8a0fd3dda9f736e5c228feb3317e9f59b0b9f9b66e461350c75bba6",
      "0xabfddf56212342693b4bbf31722ff5a32c62dc93c86b0fe01ec160b5b5cb956b",
      "0x1a91091772674b4744bd1b8ca4e6f2762bac3420c79aff1d2a9453bd845f129f",
      "0x9e941204d1b40251dcd6a75cd11a3933ceaadff951fd8e6759dabfc6ee29f43b",
      "0xee611fd0dcc2bb8389b422e37760be10d8ce2be07257bc2ef8e9d9fed576754b",
      "0xa3910064ede9ae1079e1130aba5dfcde557f062b852d03d378a48d322cc2603d",
      "0x664e15863e445b687fdbacea8e7d663c8ae5b8d0ad1eeb1d6da1ec167a1ceb17",
      "0xf1dd3335b8b85f183b19168e31f46b92860cd8e922522379286a2844fc80c31d"
    ]
  }
}`

func mustHexToBytes(hexStr string) []byte {
	bytes, err := hex.DecodeString(hexStr[2:])
	if err != nil {
		panic(err)
	}
	return bytes
}

func mustHexToBase64(hexStr string) string {
	bytes, err := hex.DecodeString(hexStr[2:])
	if err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(bytes)
}

func Test_ConstructDataResultTreeWithTestData(t *testing.T) {
	f := initFixture(t)

	var data TestData
	err := json.Unmarshal([]byte(testDataJSON), &data)
	if err != nil {
		log.Fatalf("Error unmarshalling JSON: %v", err)
	}

	// Construct DataResult objects and check their hashes.
	dataResults := make([]types.DataResult, len(data.Results))
	for i := range dataResults {
		sample := types.DataResult{
			DrId:           data.Results[i].DrID[2:],
			Version:        data.Results[i].Version,
			BlockHeight:    uint64(data.Results[i].BlockHeight),
			ExitCode:       uint32(data.Results[i].ExitCode),
			GasUsed:        uint64(data.Results[i].GasUsed),
			Result:         mustHexToBytes(data.Results[i].Result),
			PaybackAddress: mustHexToBase64(data.Results[i].PaybackAddress),
			SedaPayload:    mustHexToBase64(data.Results[i].SedaPayload),
			Consensus:      data.Results[i].Consensus,
		}

		var err error
		sample.Id, err = sample.TryHash()
		require.NoError(t, err)
		require.Equal(t, data.Results[i].ResultID[2:], sample.Id)

		dataResults[i] = sample
	}

	// Store the data results and construct the data result tree.
	for _, dr := range dataResults {
		err := f.batchingKeeper.SetDataResultForBatching(f.Context(), dr)
		require.NoError(t, err)
	}
	_, root, err := f.batchingKeeper.ConstructDataResultTree(f.Context())
	require.NoError(t, err)

	// Check the tree root (Expected root is computed assuming empty
	// previous data result root).
	expRoot := utils.RootFromLeaves([][]byte{nil, mustHexToBytes(data.Tree.Root)})
	require.Equal(t, expRoot, root)
}
