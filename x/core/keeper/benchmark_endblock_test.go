package keeper_test

// import (
// 	"encoding/base64"
// 	"fmt"
// 	"testing"

// 	"github.com/stretchr/testify/require"
// 	"golang.org/x/exp/rand"

// 	"cosmossdk.io/math"

// 	sdk "github.com/cosmos/cosmos-sdk/types"

// 	"github.com/sedaprotocol/seda-chain/testutil/testwasms"
// 	"github.com/sedaprotocol/seda-chain/x/core/types"
// )

// func BenchmarkDataRequestFlow(b *testing.B) {
// 	f := initFixture(b)

// 	proxyPubKeys := []string{"03b27f2df0cbdb5cdadff5b4be0c9fda5aa3a59557ef6d0b49b4298ef42c8ce2b0"}
// 	err := f.SetDataProxyConfig(proxyPubKeys[0], "seda1zcds6ws7l0e005h3xrmg5tx0378nyg8gtmn64f", sdk.NewCoin(bondDenom, math.NewInt(1000000000000000000)))
// 	require.NoError(b, err)

// 	params := types.DefaultParams()
// 	params.TallyConfig.MaxTalliesPerBlock = 1000
// 	f.keeper.SetParams(f.Context(), params)

// 	for i := 0; i < b.N; i++ {
// 		b.StopTimer()
// 		for j := 0; j < 1000; j++ {
// 			f.executeDataRequestFlow(
// 				b, nil, nil,
// 				1, 1, 1, false,
// 				commitRevealConfig{
// 					requestHeight: 1,
// 					requestMemo:   base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%x", rand.Int63()))),
// 					reveal:        base64.StdEncoding.EncodeToString([]byte("reveal")),
// 					proxyPubKeys:  proxyPubKeys,
// 					gasUsed:       150000000000000000,
// 				})
// 		}

// 		b.StartTimer()
// 		err = f.keeper.EndBlock(f.Context())
// 		require.NoError(b, err)

// 		f.AddBlock()
// 	}
// }

// func BenchmarkBigTallyPrograms(b *testing.B) {
// 	f := initFixture(b)

// 	params := types.DefaultParams()
// 	params.TallyConfig.MaxTalliesPerBlock = 1000
// 	f.keeper.SetParams(f.Context(), params)

// 	for i := 0; i < b.N; i++ {
// 		b.StopTimer()
// 		for j := 0; j < 1000; j++ {
// 			f.commitRevealDataRequest(
// 				b, nil, testwasms.BigWasm(),
// 				1, 1, 1, false,
// 				commitRevealConfig{
// 					requestHeight: 1,
// 					requestMemo:   base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%x", rand.Int63()))),
// 					reveal:        base64.StdEncoding.EncodeToString([]byte("reveal")),
// 					gasUsed:       150000000000000000,
// 				})
// 		}

// 		b.StartTimer()
// 		err := f.keeper.EndBlock(f.Context())
// 		require.NoError(b, err)

// 		f.AddBlock()
// 	}
// }
