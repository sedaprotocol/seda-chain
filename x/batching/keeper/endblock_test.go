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
		entriesWithSep[i] = append([]byte{utils.SEDASeparatorDataRequest}, entry...)
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
		entriesWithSep[i] = append([]byte{utils.SEDASeparatorSecp256k1}, entry.EthAddress...)
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
func addBatchSigningValidators(t *testing.T, f *fixture, num int) ([]sdk.AccAddress, [][]byte, []int64) {
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
		pubKeys[i] = elliptic.Marshal(privKey.PublicKey, privKey.PublicKey.X, privKey.PublicKey.Y)

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

		_, err = f.stakingKeeper.Keeper.EndBlocker(ctx)
		require.NoError(t, err)

		err = f.pubKeyKeeper.SetValidatorKeyAtIndex(ctx, valAddr, utils.SEDAKeyIndexSecp256k1, pubKeys[i])
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

type Validator struct {
	Identity    string `json:"identity"`
	VotingPower uint32 `json:"votingPower"`
	PrivateKey  string `json:"privateKey"`
	PublicKey   string `json:"publicKey"`
}

type Tree struct {
	Root   string   `json:"root"`
	Leaves []string `json:"leaves"`
}

type TestData struct {
	Results        []Result    `json:"results"`
	DataResultTree Tree        `json:"tree"`
	Validators     []Validator `json:"validators"`
	ValidatorTree  Tree        `json:"validatorsTree"`
}

const testDataJSON = `{
  "results": [
    {
      "resultId": "0xccf12276c43cc61e0f3c6ace3e66872eda5df5ec753525a7bddab6fa3407e927",
      "version": "0.0.1",
      "drId": "0x2e469f40cf0afb5b93641c75aaecb654f0c73fc2dd4fcc350150c4825d2fb36e",
      "consensus": true,
      "exitCode": 0,
      "result": "0x39bf027dd97f3bae0cf8cfb909695ec63313a9bd61ad52fc7f52cf565b141da8",
      "blockHeight": 0,
      "blockTimestamp": 0,
      "gasUsed": 0,
      "paybackAddress": "0x0000000000000000000000000000000000000000",
      "sedaPayload": "0x0000000000000000000000000000000000000000000000000000000000000000"
    },
    {
      "resultId": "0xe5c2d374e25002439b4d332914c1b15f438fbb8edab3c37c0c0fff4ba6f661da",
      "version": "0.0.1",
      "drId": "0xa779f7061360e9fd128acb8686c393294db943848dc6c1e2248dd91fc79f3835",
      "consensus": true,
      "exitCode": 0,
      "result": "0x39bf027dd97f3bae0cf8cfb909695ec63313a9bd61ad52fc7f52cf565b141da8",
      "blockHeight": 0,
      "blockTimestamp": 0,
      "gasUsed": 0,
      "paybackAddress": "0x0000000000000000000000000000000000000000",
      "sedaPayload": "0x0000000000000000000000000000000000000000000000000000000000000000"
    },
    {
      "resultId": "0xbc769462ef60997285d262e404bf036b0d62ebd59df10470f3c0223d1deb18c0",
      "version": "0.0.1",
      "drId": "0xcedae320cef23e6b50baf01f722b53818fd1d738d796a57a785c71436237dd35",
      "consensus": true,
      "exitCode": 0,
      "result": "0x39bf027dd97f3bae0cf8cfb909695ec63313a9bd61ad52fc7f52cf565b141da8",
      "blockHeight": 0,
      "blockTimestamp": 0,
      "gasUsed": 0,
      "paybackAddress": "0x0000000000000000000000000000000000000000",
      "sedaPayload": "0x0000000000000000000000000000000000000000000000000000000000000000"
    },
    {
      "resultId": "0xef5977dcf9789d3d51ec4741317187406636ae4f3fda398b561391d2a32ecbca",
      "version": "0.0.1",
      "drId": "0xd2e85a4f932ee01d3aec08c4fb0725425073914e11e32219a90d0fb190adcce3",
      "consensus": true,
      "exitCode": 0,
      "result": "0x39bf027dd97f3bae0cf8cfb909695ec63313a9bd61ad52fc7f52cf565b141da8",
      "blockHeight": 0,
      "blockTimestamp": 0,
      "gasUsed": 0,
      "paybackAddress": "0x0000000000000000000000000000000000000000",
      "sedaPayload": "0x0000000000000000000000000000000000000000000000000000000000000000"
    },
    {
      "resultId": "0xfb849492d85812dacb8c7a61555a642571674746253c6b36eec48c4937854198",
      "version": "0.0.1",
      "drId": "0xa11695dd090839580ee64a21412e8f0843255c149f5352311ab6960990e690be",
      "consensus": true,
      "exitCode": 0,
      "result": "0x39bf027dd97f3bae0cf8cfb909695ec63313a9bd61ad52fc7f52cf565b141da8",
      "blockHeight": 0,
      "blockTimestamp": 0,
      "gasUsed": 0,
      "paybackAddress": "0x0000000000000000000000000000000000000000",
      "sedaPayload": "0x0000000000000000000000000000000000000000000000000000000000000000"
    },
    {
      "resultId": "0x7259f83b16565218632920cc2def23a9fa6411eb10ceca5a8c2231c30adf91c0",
      "version": "0.0.1",
      "drId": "0xa4a1baa9d463128792770c0aadfc7916362d9309e19d709729060a48eb3b4b13",
      "consensus": true,
      "exitCode": 0,
      "result": "0x39bf027dd97f3bae0cf8cfb909695ec63313a9bd61ad52fc7f52cf565b141da8",
      "blockHeight": 0,
      "blockTimestamp": 0,
      "gasUsed": 0,
      "paybackAddress": "0x0000000000000000000000000000000000000000",
      "sedaPayload": "0x0000000000000000000000000000000000000000000000000000000000000000"
    },
    {
      "resultId": "0x49212d617b70d464f4c90b5ae16e89e5bc4ab57dba99a00246d4a230cdd80e4a",
      "version": "0.0.1",
      "drId": "0xe220988adad2c99912c5a4a7feb197e04a1a0396de7ca16b5e47dc1b6995b484",
      "consensus": true,
      "exitCode": 0,
      "result": "0x39bf027dd97f3bae0cf8cfb909695ec63313a9bd61ad52fc7f52cf565b141da8",
      "blockHeight": 0,
      "blockTimestamp": 0,
      "gasUsed": 0,
      "paybackAddress": "0x0000000000000000000000000000000000000000",
      "sedaPayload": "0x0000000000000000000000000000000000000000000000000000000000000000"
    },
    {
      "resultId": "0xfd9e8e35b1195457cb0fa4f0f451342821d51571eb2bd6a0edbce4f994e384f8",
      "version": "0.0.1",
      "drId": "0x2d9d50f861512a9f7041452eddfbff39f2553e1ebb5646d2eecf6733d7c8d2ce",
      "consensus": true,
      "exitCode": 0,
      "result": "0x39bf027dd97f3bae0cf8cfb909695ec63313a9bd61ad52fc7f52cf565b141da8",
      "blockHeight": 0,
      "blockTimestamp": 0,
      "gasUsed": 0,
      "paybackAddress": "0x0000000000000000000000000000000000000000",
      "sedaPayload": "0x0000000000000000000000000000000000000000000000000000000000000000"
    },
    {
      "resultId": "0x41ae733f44a6cd3c5446b9599b31c32a71fabe0d560f09929e4627b3d1ff8060",
      "version": "0.0.1",
      "drId": "0x3b8dfa0893f0a24ce3d600fad4ca16fd08ee257efea866145d000e65cf14f04a",
      "consensus": true,
      "exitCode": 0,
      "result": "0x39bf027dd97f3bae0cf8cfb909695ec63313a9bd61ad52fc7f52cf565b141da8",
      "blockHeight": 0,
      "blockTimestamp": 0,
      "gasUsed": 0,
      "paybackAddress": "0x0000000000000000000000000000000000000000",
      "sedaPayload": "0x0000000000000000000000000000000000000000000000000000000000000000"
    },
    {
      "resultId": "0x6aef851aea422b9f23c9cc1eef4e3190f968c9dcf2515d0d203670c706007e9e",
      "version": "0.0.1",
      "drId": "0xf180cde27633a009ad05fe86b0ef69c7f2065c3fcf40b103710173829990ce87",
      "consensus": true,
      "exitCode": 0,
      "result": "0x39bf027dd97f3bae0cf8cfb909695ec63313a9bd61ad52fc7f52cf565b141da8",
      "blockHeight": 0,
      "blockTimestamp": 0,
      "gasUsed": 0,
      "paybackAddress": "0x0000000000000000000000000000000000000000",
      "sedaPayload": "0x0000000000000000000000000000000000000000000000000000000000000000"
    }
  ],
  "tree": {
    "root": "0x561e1141ca38e99ad3b9d2d94744be7d6b307011b8dd35b8e08d97b4226ae32f",
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
  },
  "validators": [
    {
      "identity": "0xCc082f1F022BEA35aC8e1b24F854B36202a3028f",
      "votingPower": 10000000,
	  "privateKey": "0x2ad00ff91daf0aaea27c4f476dd4df41facab3b8387a70950850db39cf1c0426",
      "publicKey": "0x040b070bdea3df39dc6dfc2d79217318c840ef82acce1c944fa890c5ea5f22093c1d57ed4827c1d1a8c2d3bcf80a675f81b13cbf053d60f0b890d18e728ef4aaa5"
    },
    {
      "identity": "0x79492bD49B1F7B86B23C8c6405Bf1474BEd33CF9",
      "votingPower": 10000000,
	  "privateKey": "0x97b4fc537164adfe7ef340f333f4617bda558375776b7ed32c3ae403c284c669",
      "publicKey": "0x044e22b6d78452187ace54a2534dcfd7f7b057872cfcf925e4f6ccd43f328477d78f1ba4f18c3b81097402ebe09a5cb434d433489e57e0607a28f7274eee91af94"
    },
    {
      "identity": "0x1991F8B5b0cCc1B24B0C07884bEC90188f9FC07C",
      "votingPower": 10000000,
	  "privateKey": "0x813143a6e5521a0fc08d18306e47d832b55cc50cc38ae6da1adb95357adba421",
      "publicKey": "0x0474cb50c6c91ed08f6ca08e290a35a4ab1a6aaa7f04d0d3119c5732f9f72238c5743a06b32912332b48eb8471d4afcbdd61cf9b5e413d2a38571c370ffc1e67ef"
    },
    {
      "identity": "0xe0eD1759b6b7356474E310e02FD3dC8eF8c1878f",
      "votingPower": 10000000,
	  "privateKey": "0x8f329d64c288c9abef0e611c9a5808f9e524b3173c59e5641dd856c19a862dc4",
      "publicKey": "0x04fd000dacddf11c28b19825cc93f49002c51e7906027b1b489c835995dd512015e14bc4b61601d25b076c54554bf4954d4891179eceb342b24aa0e1e226c9c254"
    },
    {
      "identity": "0xD39f01574623DB56131de6C2709Ce7a8dfFAa702",
      "votingPower": 10000000,
	  "privateKey": "0x44bbdc4f0362f7349bd49d78c650c9314e9ddfc601d63b981858f195371be010",
      "publicKey": "0x045f5722e04c79ea140fe1059aee438c389047542f1f50ed4a091e632f17bb2dcc1356a0b1d74f4933c41af831ba341b9aabf81a98ad2d476879115ea8ee1a42e0"
    },
    {
      "identity": "0x6C819ddA21706b96639f39F1C930d3b74464c2e8",
      "votingPower": 10000000,
	  "privateKey": "0x5cfc4f47231ad65cfb80608d21966af3145699b766fa17bb4718a73dd3eb318f",
      "publicKey": "0x04d433d87102a16299fe3e6a1eed0aedc06062ff0a379e72190e67d76254f9991f4e40819470712d98cb2bd5df889830023bc3dcd40c80eac02b5249f4d14b1490"
    },
    {
      "identity": "0xF144e8ddE082dB431B24BdB442Cf38Cf365E256C",
      "votingPower": 10000000,
	  "privateKey": "0x9eafdf194ea0e55fe4fda6666f75d16825154c65ebde0d755a951a5e4ec10577",
      "publicKey": "0x04b1be49abfe20e6f556a53e0222a55e8901961ecc2ee28e4c5338c149d130395a2fd1684067c29b620a802fae8d5e4cc935cdc59da9b3b32e84c124841cde37a2"
    },
    {
      "identity": "0x1Dc65448F3Feb1BdA46267BCB6b87dA4ac601217",
      "votingPower": 10000000,
	  "privateKey": "0x946b352930a304d8af904739adbbd91ba750d76d8c80c5d0d349b03cede5f5cb",
      "publicKey": "0x04df13483dd120b1dc572d5d02579d8d3a529fb01efdb7e5863088fdcd5b081787d67bccc01d8902075173b6ab845f327b90a04dd7f82cd95b9b6d81c8ad9d3fce"
    },
    {
      "identity": "0x0dE605f6e31d27F4Ca75a5ac2A396237791A394B",
      "votingPower": 10000000,
	  "privateKey": "0x9dac23b2a961fc86b7d8a1f1c0d20634b5b9af9d212a3b3895f3331afad1fefd",
      "publicKey": "0x046ac264122705205bf5db9f5d2e15d929ba97d3efd1917b37bdc4e046c3b433080a1d69e2a8ae266f98123784a503a6aeabe3fe71645d75cf86bd52e6e8426b51"
    },
    {
      "identity": "0x90599B1969C5CF8A34240a2C8A7a823E8eb1f395",
      "votingPower": 10000000,
	  "privateKey": "0xc3ff80fc4a3f3c0941aad43ee71675536481e6d04338f5356ed644bb99c5df80",
      "publicKey": "0x04f4bb0dd9b3d60988aa8e983be59f3b04ad55925ca357deb882b970177cd22d82e3b12c0adce5c86aa4833f8ae477dfbaeba4b1312a3c2beefbb1ace48f0b43aa"
    }
  ],
  "validatorsTree": {
    "root": "0x2c5073e9c4308e65eb22152f62611c6f604d6b427c6514bbd1effaf282d9614a",
    "leaves": [
      "0xfd9e8e35b1195457cb0fa4f0f451342821d51571eb2bd6a0edbce4f994e384f8",
      "0xccf12276c43cc61e0f3c6ace3e66872eda5df5ec753525a7bddab6fa3407e927",
      "0x41ae733f44a6cd3c5446b9599b31c32a71fabe0d560f09929e4627b3d1ff8060",
      "0xfb849492d85812dacb8c7a61555a642571674746253c6b36eec48c4937854198",
      "0x7259f83b16565218632920cc2def23a9fa6411eb10ceca5a8c2231c30adf91c0",
      "0xe5c2d374e25002439b4d332914c1b15f438fbb8edab3c37c0c0fff4ba6f661da",
      "0xbc769462ef60997285d262e404bf036b0d62ebd59df10470f3c0223d1deb18c0",
      "0xef5977dcf9789d3d51ec4741317187406636ae4f3fda398b561391d2a32ecbca",
      "0x49212d617b70d464f4c90b5ae16e89e5bc4ab57dba99a00246d4a230cdd80e4a",
      "0x6aef851aea422b9f23c9cc1eef4e3190f968c9dcf2515d0d203670c706007e9e"
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

	_, _, powers := addBatchSigningValidatorsFromTestData(t, f, data.Validators)

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

func addBatchSigningValidatorsFromTestData(t *testing.T, f *fixture, testData []Validator) ([]sdk.AccAddress, [][]byte, []int64) {
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

		_, err = f.stakingKeeper.Keeper.EndBlocker(ctx)
		require.NoError(t, err)

		privKey, err := ethcrypto.HexToECDSA(testData[i].PrivateKey[2:])
		require.NoError(t, err)
		pk := elliptic.Marshal(privKey.PublicKey, privKey.PublicKey.X, privKey.PublicKey.Y)

		err = f.pubKeyKeeper.SetValidatorKeyAtIndex(ctx, valAddr, utils.SEDAKeyIndexSecp256k1, pk)
		require.NoError(t, err)
	}
	return addrs, secp256k1PubKeys, powers
}
