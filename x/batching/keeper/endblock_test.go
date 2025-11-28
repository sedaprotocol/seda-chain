package keeper_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	cryptorand "crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"testing"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkstakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	sdkstakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/sedaprotocol/seda-chain/app/utils"
	sedatypes "github.com/sedaprotocol/seda-chain/types"
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
	entriesWithSep := make([][]byte, len(entries.Entries)) // add domain separators for tree re-construction
	for i, entry := range entries.Entries {
		entryHexes = append(entryHexes, hex.EncodeToString(entry))
		drIds = append(drIds, dataResults[i].Id)
		entriesWithSep[i] = append([]byte{sedatypes.SEDASeparatorDataResult}, entry...)
	}
	require.ElementsMatch(t, drIds, entryHexes)

	// Generate proof for each entry and verify.
	require.NoError(t, err)
	for i := range entriesWithSep {
		proof, err := utils.GetProof(entriesWithSep, i)
		require.NoError(t, err)

		ret := utils.VerifyProof(proof, root, entriesWithSep[i])
		require.True(t, ret)
	}
}

func Test_ConstructValidatorTree(t *testing.T) {
	f := initFixture(t)
	_, pks, powers := f.addBatchSigningValidators(t, 10)

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
	expectedAddrs := make([][]byte, len(entries))
	parsedAddrs := make([][]byte, len(entries))
	parsedPowers := make([]uint32, len(entries))
	entriesWithSep := make([][]byte, len(entries))
	for i, entry := range entries {
		parsedAddrs[i] = entry.EthAddress
		parsedPowers[i] = entry.VotingPowerPercent
		expectedAddrs[i], err = utils.PubKeyToEthAddress(pks[i])
		require.NoError(t, err)

		// Reconstruct the validator tree entry.
		entriesWithSep[i] = append([]byte{sedatypes.SEDASeparatorSecp256k1}, entry.EthAddress...)
		entriesWithSep[i] = binary.BigEndian.AppendUint32(entriesWithSep[i], entry.VotingPowerPercent)
	}
	require.ElementsMatch(t, expectedAddrs, parsedAddrs)
	require.ElementsMatch(t, powerPercents, parsedPowers)

	// Generate proof for each entry and verify.
	require.NoError(t, err)
	for i := range entriesWithSep {
		pf, err := utils.GetProof(entriesWithSep, i)
		require.NoError(t, err)
		ret := utils.VerifyProof(pf, root, entriesWithSep[i])
		require.True(t, ret)
	}
}

// addBatchSigningValidators funds test addresses, adds them as validators,
// and registers secp256k1 public keys for their batch signing.
func (f *fixture) addBatchSigningValidators(t testing.TB, num int) ([]sdk.AccAddress, [][]byte, []int64) {
	t.Helper()

	ctx := f.Context()
	stakingMsgSvr := sdkstakingkeeper.NewMsgServerImpl(f.stakingKeeper.Keeper)

	addrs := simtestutil.AddTestAddrs(f.bankKeeper, f.stakingKeeper, ctx, num, math.NewIntFromUint64(10000000000000000000))
	pubKeys := make([][]byte, len(addrs))
	powers := make([]int64, len(addrs))
	for i, addr := range addrs {
		valAddr := sdk.ValAddress(addr)
		valPubKey := secp256k1.GenPrivKey().PubKey()
		powers[i] = int64(rand.Intn(10) + 1)

		privKey, err := ecdsa.GenerateKey(ethcrypto.S256(), cryptorand.Reader)
		if err != nil {
			panic(fmt.Sprintf("failed to generate secp256k1 private key: %v", err))
		}
		pubKeys[i] = elliptic.Marshal(privKey.PublicKey, privKey.X, privKey.Y)

		valTokens := sdk.TokensFromConsensusPower(powers[i], sdk.DefaultPowerReduction)
		valCreateMsg, err := sdkstakingtypes.NewMsgCreateValidator(
			valAddr.String(),
			valPubKey,
			sdk.NewCoin(bondDenom, valTokens),
			sdkstakingtypes.NewDescription("T", "E", "S", "T", "Z"),
			sdkstakingtypes.NewCommissionRates(math.LegacyZeroDec(), math.LegacyZeroDec(), math.LegacyZeroDec()),
			math.OneInt(),
		)
		require.NoError(t, err)
		res, err := stakingMsgSvr.CreateValidator(ctx, valCreateMsg)
		require.NoError(t, err)
		require.NotNil(t, res)

		_, err = f.stakingKeeper.EndBlocker(ctx)
		require.NoError(t, err)

		err = f.pubKeyKeeper.SetValidatorKeyAtIndex(ctx, valAddr, sedatypes.SEDAKeyIndexSecp256k1, pubKeys[i])
		require.NoError(t, err)
	}
	return addrs, pubKeys, powers
}

// generateDataResults returns a given number of randomly-generated
// data results.
func generateDataResults(t testing.TB, num int) []types.DataResult {
	t.Helper()
	dataResults := make([]types.DataResult, num)
	for i := range dataResults {
		gasUsed := math.NewIntFromUint64(rand.Uint64())
		sample := types.DataResult{
			DrId:           generateRandomHexString(64),
			Version:        fmt.Sprintf("%d.%d.%d", rand.Intn(10), rand.Intn(10), rand.Intn(10)),
			BlockHeight:    rand.Uint64(),
			ExitCode:       rand.Uint32(),
			GasUsed:        &gasUsed,
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
	BlockTimestamp int    `json:"blockTimestamp"`
	GasUsed        string `json:"gasUsed"`
	PaybackAddress string `json:"paybackAddress"`
	SedaPayload    string `json:"sedaPayload"`
}

type Validator struct {
	Identity    string `json:"identity"`
	VotingPower uint32 `json:"votingPower"`
}

type Tree struct {
	Root   string   `json:"root"`
	Leaves []string `json:"leaves"`
}

type Wallet struct {
	Address    string `json:"address"`
	PrivateKey string `json:"privateKey"`
	PublicKey  string `json:"publicKey"`
}

type TestData struct {
	Results        []Result    `json:"results"`
	DataResultTree Tree        `json:"resultsTree"`
	Validators     []Validator `json:"validators"`
	ValidatorTree  Tree        `json:"validatorsTree"`
	Wallets        []Wallet    `json:"wallets"`
}

const testDataJSON = `{
  "requests": [
    {
      "requestId": "0x341f8933b3a10b929865d71d2f575794e0fbbafb7801b87975ebc7df20523e79",
      "execProgramId": "0x0000000000000000000000000000000000000000000000000000000000000000",
      "execInputs": "0x",
      "execGasLimit": "1000000",
      "tallyProgramId": "0x0000000000000000000000000000000000000000000000000000000000000000",
      "tallyInputs": "0x",
      "tallyGasLimit": "1000000",
      "replicationFactor": 1,
      "consensusFilter": "0x00",
      "gasPrice": "0.0",
      "memo": "0x00"
    },
    {
      "requestId": "0x71d862d090d5b0cfc0e38c553c42efff3f4f1fe17440104df2a6ed228eb1e4b2",
      "execProgramId": "0x0000000000000000000000000000000000000000000000000000000000000000",
      "execInputs": "0x",
      "execGasLimit": "1000000",
      "tallyProgramId": "0x0000000000000000000000000000000000000000000000000000000000000000",
      "tallyInputs": "0x",
      "tallyGasLimit": "1000000",
      "replicationFactor": 1,
      "consensusFilter": "0x00",
      "gasPrice": "0.0",
      "memo": "0x01"
    },
    {
      "requestId": "0xc320b7f89d04f6911165f36ae7d3b87a34e8fb6b41f7fc7dedeb49dcadda4e18",
      "execProgramId": "0x0000000000000000000000000000000000000000000000000000000000000000",
      "execInputs": "0x",
      "execGasLimit": "1000000",
      "tallyProgramId": "0x0000000000000000000000000000000000000000000000000000000000000000",
      "tallyInputs": "0x",
      "tallyGasLimit": "1000000",
      "replicationFactor": 1,
      "consensusFilter": "0x00",
      "gasPrice": "0.0",
      "memo": "0x02"
    },
    {
      "requestId": "0xfdbe500a36237c6d41a8d1d75c901f8f308ba460b7635ec649eae820e70818f1",
      "execProgramId": "0x0000000000000000000000000000000000000000000000000000000000000000",
      "execInputs": "0x",
      "execGasLimit": "1000000",
      "tallyProgramId": "0x0000000000000000000000000000000000000000000000000000000000000000",
      "tallyInputs": "0x",
      "tallyGasLimit": "1000000",
      "replicationFactor": 1,
      "consensusFilter": "0x00",
      "gasPrice": "0.0",
      "memo": "0x03"
    },
    {
      "requestId": "0x5e9dc95872266e95a668f2cd82eb5316076f4c6828f824f5662673598838f8cf",
      "execProgramId": "0x0000000000000000000000000000000000000000000000000000000000000000",
      "execInputs": "0x",
      "execGasLimit": "1000000",
      "tallyProgramId": "0x0000000000000000000000000000000000000000000000000000000000000000",
      "tallyInputs": "0x",
      "tallyGasLimit": "1000000",
      "replicationFactor": 1,
      "consensusFilter": "0x00",
      "gasPrice": "0.0",
      "memo": "0x04"
    },
    {
      "requestId": "0xff7ec2a6f8d2c300cfb9643a9d5733069716671906a0088a05feea3ffa878cd9",
      "execProgramId": "0x0000000000000000000000000000000000000000000000000000000000000000",
      "execInputs": "0x",
      "execGasLimit": "1000000",
      "tallyProgramId": "0x0000000000000000000000000000000000000000000000000000000000000000",
      "tallyInputs": "0x",
      "tallyGasLimit": "1000000",
      "replicationFactor": 1,
      "consensusFilter": "0x00",
      "gasPrice": "0.0",
      "memo": "0x05"
    },
    {
      "requestId": "0x2f1a4631e4f44f75b327b05bc470685d35f81e38b2be215df912c0c733516d60",
      "execProgramId": "0x0000000000000000000000000000000000000000000000000000000000000000",
      "execInputs": "0x",
      "execGasLimit": "1000000",
      "tallyProgramId": "0x0000000000000000000000000000000000000000000000000000000000000000",
      "tallyInputs": "0x",
      "tallyGasLimit": "1000000",
      "replicationFactor": 1,
      "consensusFilter": "0x00",
      "gasPrice": "0.0",
      "memo": "0x06"
    },
    {
      "requestId": "0xa42faa4d3136c37ae60c76f2e648a4d3aec0865b1b9fcf688fcbe9caf8007126",
      "execProgramId": "0x0000000000000000000000000000000000000000000000000000000000000000",
      "execInputs": "0x",
      "execGasLimit": "1000000",
      "tallyProgramId": "0x0000000000000000000000000000000000000000000000000000000000000000",
      "tallyInputs": "0x",
      "tallyGasLimit": "1000000",
      "replicationFactor": 1,
      "consensusFilter": "0x00",
      "gasPrice": "0.0",
      "memo": "0x07"
    },
    {
      "requestId": "0xddddc03d467cb94bcea473231c992cf66678e59645ab906802f8e06891dd9d2d",
      "execProgramId": "0x0000000000000000000000000000000000000000000000000000000000000000",
      "execInputs": "0x",
      "execGasLimit": "1000000",
      "tallyProgramId": "0x0000000000000000000000000000000000000000000000000000000000000000",
      "tallyInputs": "0x",
      "tallyGasLimit": "1000000",
      "replicationFactor": 1,
      "consensusFilter": "0x00",
      "gasPrice": "0.0",
      "memo": "0x08"
    },
    {
      "requestId": "0x19888f9d7530c6e5db252d6caa112939bbbfa6f1a634311152ffeb2b0e2b45d4",
      "execProgramId": "0x0000000000000000000000000000000000000000000000000000000000000000",
      "execInputs": "0x",
      "execGasLimit": "1000000",
      "tallyProgramId": "0x0000000000000000000000000000000000000000000000000000000000000000",
      "tallyInputs": "0x",
      "tallyGasLimit": "1000000",
      "replicationFactor": 1,
      "consensusFilter": "0x00",
      "gasPrice": "0.0",
      "memo": "0x09"
    }
  ],
  "results": [
    {
      "resultId": "0xd0bf1769908fbd8927d70db3067b2f1f81b997cf534cce91d16c83d1272422d6",
      "version": "0.0.1",
      "drId": "0x341f8933b3a10b929865d71d2f575794e0fbbafb7801b87975ebc7df20523e79",
      "consensus": true,
      "exitCode": 0,
      "result": "0x39bf027dd97f3bae0cf8cfb909695ec63313a9bd61ad52fc7f52cf565b141da8",
      "blockHeight": 0,
      "blockTimestamp": 1737473263,
      "gasUsed": "0",
      "paybackAddress": "0x0000000000000000000000000000000000000000",
      "sedaPayload": "0x0000000000000000000000000000000000000000000000000000000000000000"
    },
    {
      "resultId": "0x905b8931c258204cab4bfa07b3662dc5ee5b39e5142a715c89a3d70715220d04",
      "version": "0.0.1",
      "drId": "0x71d862d090d5b0cfc0e38c553c42efff3f4f1fe17440104df2a6ed228eb1e4b2",
      "consensus": true,
      "exitCode": 0,
      "result": "0x39bf027dd97f3bae0cf8cfb909695ec63313a9bd61ad52fc7f52cf565b141da8",
      "blockHeight": 0,
      "blockTimestamp": 1737473263,
      "gasUsed": "0",
      "paybackAddress": "0x0000000000000000000000000000000000000000",
      "sedaPayload": "0x0000000000000000000000000000000000000000000000000000000000000000"
    },
    {
      "resultId": "0x0e5a051af3e9c25e640c2ef8313e46329c71b31a16e4cf80c80b145251e8ca64",
      "version": "0.0.1",
      "drId": "0xc320b7f89d04f6911165f36ae7d3b87a34e8fb6b41f7fc7dedeb49dcadda4e18",
      "consensus": true,
      "exitCode": 0,
      "result": "0x39bf027dd97f3bae0cf8cfb909695ec63313a9bd61ad52fc7f52cf565b141da8",
      "blockHeight": 0,
      "blockTimestamp": 1737473263,
      "gasUsed": "0",
      "paybackAddress": "0x0000000000000000000000000000000000000000",
      "sedaPayload": "0x0000000000000000000000000000000000000000000000000000000000000000"
    },
    {
      "resultId": "0xc2ca9464d7c33d3c1c3dc2eaf8e76791bf42b37e207a91fdf460cf3eb32f5dc1",
      "version": "0.0.1",
      "drId": "0xfdbe500a36237c6d41a8d1d75c901f8f308ba460b7635ec649eae820e70818f1",
      "consensus": true,
      "exitCode": 0,
      "result": "0x39bf027dd97f3bae0cf8cfb909695ec63313a9bd61ad52fc7f52cf565b141da8",
      "blockHeight": 0,
      "blockTimestamp": 1737473263,
      "gasUsed": "0",
      "paybackAddress": "0x0000000000000000000000000000000000000000",
      "sedaPayload": "0x0000000000000000000000000000000000000000000000000000000000000000"
    },
    {
      "resultId": "0xb817a376b69ce89ae8ac0c54afcb34e0e1ed82a034ad5efff76df36d88942e62",
      "version": "0.0.1",
      "drId": "0x5e9dc95872266e95a668f2cd82eb5316076f4c6828f824f5662673598838f8cf",
      "consensus": true,
      "exitCode": 0,
      "result": "0x39bf027dd97f3bae0cf8cfb909695ec63313a9bd61ad52fc7f52cf565b141da8",
      "blockHeight": 0,
      "blockTimestamp": 1737473263,
      "gasUsed": "0",
      "paybackAddress": "0x0000000000000000000000000000000000000000",
      "sedaPayload": "0x0000000000000000000000000000000000000000000000000000000000000000"
    },
    {
      "resultId": "0xf41e386ffbcc7890d5c657aa3b26160048a8942f481e9bee52a5c5abcb68d1c9",
      "version": "0.0.1",
      "drId": "0xff7ec2a6f8d2c300cfb9643a9d5733069716671906a0088a05feea3ffa878cd9",
      "consensus": true,
      "exitCode": 0,
      "result": "0x39bf027dd97f3bae0cf8cfb909695ec63313a9bd61ad52fc7f52cf565b141da8",
      "blockHeight": 0,
      "blockTimestamp": 1737473263,
      "gasUsed": "0",
      "paybackAddress": "0x0000000000000000000000000000000000000000",
      "sedaPayload": "0x0000000000000000000000000000000000000000000000000000000000000000"
    },
    {
      "resultId": "0x22b561e9981c743dc0426464406e867d5aeb38043834f5fd92de464592c93db9",
      "version": "0.0.1",
      "drId": "0x2f1a4631e4f44f75b327b05bc470685d35f81e38b2be215df912c0c733516d60",
      "consensus": true,
      "exitCode": 0,
      "result": "0x39bf027dd97f3bae0cf8cfb909695ec63313a9bd61ad52fc7f52cf565b141da8",
      "blockHeight": 0,
      "blockTimestamp": 1737473263,
      "gasUsed": "0",
      "paybackAddress": "0x0000000000000000000000000000000000000000",
      "sedaPayload": "0x0000000000000000000000000000000000000000000000000000000000000000"
    },
    {
      "resultId": "0xd0fc776a29dc7799815462385827151946535943e2a5c7fdc3b7a26dd97d82e2",
      "version": "0.0.1",
      "drId": "0xa42faa4d3136c37ae60c76f2e648a4d3aec0865b1b9fcf688fcbe9caf8007126",
      "consensus": true,
      "exitCode": 0,
      "result": "0x39bf027dd97f3bae0cf8cfb909695ec63313a9bd61ad52fc7f52cf565b141da8",
      "blockHeight": 0,
      "blockTimestamp": 1737473263,
      "gasUsed": "0",
      "paybackAddress": "0x0000000000000000000000000000000000000000",
      "sedaPayload": "0x0000000000000000000000000000000000000000000000000000000000000000"
    },
    {
      "resultId": "0xbe9bceb076932fc3cd6267a126fe2f9ffc0e1c3187348710b2675f829da45613",
      "version": "0.0.1",
      "drId": "0xddddc03d467cb94bcea473231c992cf66678e59645ab906802f8e06891dd9d2d",
      "consensus": true,
      "exitCode": 0,
      "result": "0x39bf027dd97f3bae0cf8cfb909695ec63313a9bd61ad52fc7f52cf565b141da8",
      "blockHeight": 0,
      "blockTimestamp": 1737473263,
      "gasUsed": "0",
      "paybackAddress": "0x0000000000000000000000000000000000000000",
      "sedaPayload": "0x0000000000000000000000000000000000000000000000000000000000000000"
    },
    {
      "resultId": "0x68966ac55c98ab6d57912d351d64f01063f196aa9f53d9da5563a41eb86860a1",
      "version": "0.0.1",
      "drId": "0x19888f9d7530c6e5db252d6caa112939bbbfa6f1a634311152ffeb2b0e2b45d4",
      "consensus": true,
      "exitCode": 0,
      "result": "0x39bf027dd97f3bae0cf8cfb909695ec63313a9bd61ad52fc7f52cf565b141da8",
      "blockHeight": 0,
      "blockTimestamp": 1737473263,
      "gasUsed": "0",
      "paybackAddress": "0x0000000000000000000000000000000000000000",
      "sedaPayload": "0x0000000000000000000000000000000000000000000000000000000000000000"
    }
  ],
  "resultsTree": {
    "root": "0xde27b7ec042e0a90db95bab976037b3e3f27f074c4d712e9e67b840e420d43e2",
    "leaves": [
      "0xc643579d2b43a202ce65840dbe505c423ff4d01df3410396a852054de8ff7c99",
      "0xe478a7f8a1d20a2d7200edfe6b8bd23b7fd01f83f584e814ce4cbe912e85a76e",
      "0x9c980071053a80f06a52c65373380c5ad9fc00ae69c960a1ffe4c22d1dbc68fe",
      "0xc4b91d44bddd1fb334bae26436a41d61ac85c09326d0c01194b1a99911a12702",
      "0x9556b5f040a97599cc180acc9b4aeb08b721065525d3be2e2a85fa40b0d27cc3",
      "0x35ddbed12915089229396fb173526c23a6c5663f657849406e7451d1b3999c8b",
      "0x1cac575316f3ff2cfa4d1c168528e84fb05b3c2f4b1681438074ecaca01b1786",
      "0xda3c0b07d78808a268df695c8002bce22ac50c64c23f4f1177098c4601e72eaf",
      "0x24185e15a4391669e18145f7df2e0ebb9ceba283dd6749c13a7b0ec08dbb0aae",
      "0xab7784227f5d0ad581ad01d9ef71e92226ea442c7ce59eeef7f3fe9d59ddd9a8"
    ]
  },
  "validators": [
    {
      "identity": "0xCc082f1F022BEA35aC8e1b24F854B36202a3028f",
      "votingPower": 10000000
    },
    {
      "identity": "0x79492bD49B1F7B86B23C8c6405Bf1474BEd33CF9",
      "votingPower": 10000000
    },
    {
      "identity": "0x1991F8B5b0cCc1B24B0C07884bEC90188f9FC07C",
      "votingPower": 10000000
    },
    {
      "identity": "0xe0eD1759b6b7356474E310e02FD3dC8eF8c1878f",
      "votingPower": 10000000
    },
    {
      "identity": "0xD39f01574623DB56131de6C2709Ce7a8dfFAa702",
      "votingPower": 10000000
    },
    {
      "identity": "0x6C819ddA21706b96639f39F1C930d3b74464c2e8",
      "votingPower": 10000000
    },
    {
      "identity": "0xF144e8ddE082dB431B24BdB442Cf38Cf365E256C",
      "votingPower": 10000000
    },
    {
      "identity": "0x1Dc65448F3Feb1BdA46267BCB6b87dA4ac601217",
      "votingPower": 10000000
    },
    {
      "identity": "0x0dE605f6e31d27F4Ca75a5ac2A396237791A394B",
      "votingPower": 10000000
    },
    {
      "identity": "0x90599B1969C5CF8A34240a2C8A7a823E8eb1f395",
      "votingPower": 10000000
    }
  ],
  "validatorsTree": {
    "root": "0x2c5073e9c4308e65eb22152f62611c6f604d6b427c6514bbd1effaf282d9614a",
    "leaves": [
      "0x38519bc19a6c21b2e4d5c07f5c317a04907d74428e84ce53d7e31660363697e3",
      "0xefaaa7508118895a0d04b782da3824d32e8068bc149eb9cb64d5035bfdc23d75",
      "0x08acd70e8df30e046db9b1f428e801b98426493443630c1a10ed41ca831b59dc",
      "0xb83b5ecaf184b6525c995e98885c3c26105981d368fba6a32f118d963a8b2f0e",
      "0x5ea122e40dd0c50ea173ed35af483f3391d910051d0b81b5764962bc7259c75a",
      "0x1b36b415df57e691bc923f10a6dc481caf1b31a77f3514c443e78e07a9692e70",
      "0x15c3bea022ed633d5aeabfeba41955d19f23ad533c95d50b121b1cb0c585c2f7",
      "0xe2f6ad28acf622aad072b4af09ff3e4588c312461fd9f34f87a30398c4874c6f",
      "0x66931389acd9b85bf84eb1b4b391e0100affed572a4b5695bcbb07c1655fbfe7",
      "0x49f4e02c70dc0cff353603c5af12911e3fe408e8263005cbe2877e2677789469"
    ]
  },
  "wallets": [
    {
      "address": "0xCc082f1F022BEA35aC8e1b24F854B36202a3028f",
      "privateKey": "0x2ad00ff91daf0aaea27c4f476dd4df41facab3b8387a70950850db39cf1c0426",
      "publicKey": "0x040b070bdea3df39dc6dfc2d79217318c840ef82acce1c944fa890c5ea5f22093c1d57ed4827c1d1a8c2d3bcf80a675f81b13cbf053d60f0b890d18e728ef4aaa5"
    },
    {
      "address": "0x79492bD49B1F7B86B23C8c6405Bf1474BEd33CF9",
      "privateKey": "0x97b4fc537164adfe7ef340f333f4617bda558375776b7ed32c3ae403c284c669",
      "publicKey": "0x044e22b6d78452187ace54a2534dcfd7f7b057872cfcf925e4f6ccd43f328477d78f1ba4f18c3b81097402ebe09a5cb434d433489e57e0607a28f7274eee91af94"
    },
    {
      "address": "0x1991F8B5b0cCc1B24B0C07884bEC90188f9FC07C",
      "privateKey": "0x813143a6e5521a0fc08d18306e47d832b55cc50cc38ae6da1adb95357adba421",
      "publicKey": "0x0474cb50c6c91ed08f6ca08e290a35a4ab1a6aaa7f04d0d3119c5732f9f72238c5743a06b32912332b48eb8471d4afcbdd61cf9b5e413d2a38571c370ffc1e67ef"
    },
    {
      "address": "0xe0eD1759b6b7356474E310e02FD3dC8eF8c1878f",
      "privateKey": "0x8f329d64c288c9abef0e611c9a5808f9e524b3173c59e5641dd856c19a862dc4",
      "publicKey": "0x04fd000dacddf11c28b19825cc93f49002c51e7906027b1b489c835995dd512015e14bc4b61601d25b076c54554bf4954d4891179eceb342b24aa0e1e226c9c254"
    },
    {
      "address": "0xD39f01574623DB56131de6C2709Ce7a8dfFAa702",
      "privateKey": "0x44bbdc4f0362f7349bd49d78c650c9314e9ddfc601d63b981858f195371be010",
      "publicKey": "0x045f5722e04c79ea140fe1059aee438c389047542f1f50ed4a091e632f17bb2dcc1356a0b1d74f4933c41af831ba341b9aabf81a98ad2d476879115ea8ee1a42e0"
    },
    {
      "address": "0x6C819ddA21706b96639f39F1C930d3b74464c2e8",
      "privateKey": "0x5cfc4f47231ad65cfb80608d21966af3145699b766fa17bb4718a73dd3eb318f",
      "publicKey": "0x04d433d87102a16299fe3e6a1eed0aedc06062ff0a379e72190e67d76254f9991f4e40819470712d98cb2bd5df889830023bc3dcd40c80eac02b5249f4d14b1490"
    },
    {
      "address": "0xF144e8ddE082dB431B24BdB442Cf38Cf365E256C",
      "privateKey": "0x9eafdf194ea0e55fe4fda6666f75d16825154c65ebde0d755a951a5e4ec10577",
      "publicKey": "0x04b1be49abfe20e6f556a53e0222a55e8901961ecc2ee28e4c5338c149d130395a2fd1684067c29b620a802fae8d5e4cc935cdc59da9b3b32e84c124841cde37a2"
    },
    {
      "address": "0x1Dc65448F3Feb1BdA46267BCB6b87dA4ac601217",
      "privateKey": "0x946b352930a304d8af904739adbbd91ba750d76d8c80c5d0d349b03cede5f5cb",
      "publicKey": "0x04df13483dd120b1dc572d5d02579d8d3a529fb01efdb7e5863088fdcd5b081787d67bccc01d8902075173b6ab845f327b90a04dd7f82cd95b9b6d81c8ad9d3fce"
    },
    {
      "address": "0x0dE605f6e31d27F4Ca75a5ac2A396237791A394B",
      "privateKey": "0x9dac23b2a961fc86b7d8a1f1c0d20634b5b9af9d212a3b3895f3331afad1fefd",
      "publicKey": "0x046ac264122705205bf5db9f5d2e15d929ba97d3efd1917b37bdc4e046c3b433080a1d69e2a8ae266f98123784a503a6aeabe3fe71645d75cf86bd52e6e8426b51"
    },
    {
      "address": "0x90599B1969C5CF8A34240a2C8A7a823E8eb1f395",
      "privateKey": "0xc3ff80fc4a3f3c0941aad43ee71675536481e6d04338f5356ed644bb99c5df80",
      "publicKey": "0x04f4bb0dd9b3d60988aa8e983be59f3b04ad55925ca357deb882b970177cd22d82e3b12c0adce5c86aa4833f8ae477dfbaeba4b1312a3c2beefbb1ace48f0b43aa"
    }
  ]
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
		gasUsed, ok := math.NewIntFromString(data.Results[i].GasUsed)
		require.True(t, ok)

		sample := types.DataResult{
			DrId:           data.Results[i].DrID[2:],
			Version:        data.Results[i].Version,
			BlockHeight:    uint64(data.Results[i].BlockHeight),
			BlockTimestamp: uint64(data.Results[i].BlockTimestamp),
			ExitCode:       uint32(data.Results[i].ExitCode),
			GasUsed:        &gasUsed,
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
	_, root, err := f.batchingKeeper.ConstructDataResultTree(f.Context(), rand.Uint64())
	require.NoError(t, err)

	// Check the tree root (Expected root is computed assuming empty
	// previous data result root).
	require.Equal(t, mustHexToBytes(data.DataResultTree.Root), root)
}

func Test_ConstructValidatorTreeWithTestData(t *testing.T) {
	f := initFixture(t)

	var data TestData
	err := json.Unmarshal([]byte(testDataJSON), &data)
	if err != nil {
		log.Fatalf("Error unmarshalling JSON: %v", err)
	}

	_, _, powers := f.addBatchSigningValidatorsFromTestData(t, data.Validators, data.Wallets)

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
	expectedAddrs := make([][]byte, len(entries))
	parsedAddrs := make([][]byte, len(entries))
	parsedPowers := make([]uint32, len(entries))
	for i, entry := range entries {
		parsedAddrs[i] = entry.EthAddress
		parsedPowers[i] = entry.VotingPowerPercent
		expectedAddr, err := hex.DecodeString(data.Validators[i].Identity[2:])
		require.NoError(t, err)
		expectedAddrs[i] = expectedAddr
	}
	require.ElementsMatch(t, expectedAddrs, parsedAddrs)
	require.ElementsMatch(t, powerPercents, parsedPowers)

	// Check the tree root (Expected root is computed assuming empty
	// previous data result root).
	require.Equal(t, mustHexToBytes(data.ValidatorTree.Root), root)
}

func (f *fixture) addBatchSigningValidatorsFromTestData(t *testing.T, testData []Validator, wallets []Wallet) ([]sdk.AccAddress, [][]byte, []int64) {
	t.Helper()

	ctx := f.Context()
	stakingMsgSvr := sdkstakingkeeper.NewMsgServerImpl(f.stakingKeeper.Keeper)

	num := len(testData)
	accAmt, ok := math.NewIntFromString("10000000000000000000000000")
	require.True(t, ok)
	addrs := simtestutil.AddTestAddrs(f.bankKeeper, f.stakingKeeper, ctx, num, accAmt)
	secp256k1PubKeys := make([][]byte, len(addrs))
	powers := make([]int64, len(addrs))
	for i, addr := range addrs {
		valAddr := sdk.ValAddress(addr)
		valPubKey := secp256k1.GenPrivKey().PubKey()
		powers[i] = int64(testData[i].VotingPower)

		valTokens := sdk.TokensFromConsensusPower(powers[i], sdk.DefaultPowerReduction)
		valCreateMsg, err := sdkstakingtypes.NewMsgCreateValidator(
			valAddr.String(),
			valPubKey,
			sdk.NewCoin(bondDenom, valTokens),
			sdkstakingtypes.NewDescription("T", "E", "S", "T", "Z"),
			sdkstakingtypes.NewCommissionRates(math.LegacyZeroDec(), math.LegacyZeroDec(), math.LegacyZeroDec()),
			math.OneInt(),
		)
		require.NoError(t, err)
		res, err := stakingMsgSvr.CreateValidator(ctx, valCreateMsg)
		require.NoError(t, err)
		require.NotNil(t, res)

		_, err = f.stakingKeeper.EndBlocker(ctx)
		require.NoError(t, err)

		// We assume that the validator index is the same as the wallet index.
		privKey, err := ethcrypto.HexToECDSA(wallets[i].PrivateKey[2:])
		require.NoError(t, err)
		pk := elliptic.Marshal(privKey.PublicKey, privKey.X, privKey.Y)

		err = f.pubKeyKeeper.SetValidatorKeyAtIndex(ctx, valAddr, sedatypes.SEDAKeyIndexSecp256k1, pk)
		require.NoError(t, err)
	}
	return addrs, secp256k1PubKeys, powers
}
