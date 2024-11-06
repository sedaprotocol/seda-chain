package fuzz

import (
	"os"
	"testing"
	"time"

	cosmossdk_io_math "cosmossdk.io/math"
	"github.com/cometbft/cometbft/crypto/secp256k1"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"

	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	simcli "github.com/cosmos/cosmos-sdk/x/simulation/client/cli"

	"github.com/sedaprotocol/seda-chain/app"
	dataproxytypes "github.com/sedaprotocol/seda-chain/x/data-proxy/types"

	"github.com/stretchr/testify/require"
)

func init() {
	simcli.GetSimulatorFlags()
}

func FuzzSetupSim(f *testing.F) {
	f.Fuzz(func(t *testing.T, seed int64, chainID string) {
		simcli.FlagSeedValue = seed
		simcli.FlagVerboseValue = true
		simcli.FlagCommitValue = true
		simcli.FlagEnabledValue = true

		config := simcli.NewConfigFromFlags()
		config.ChainID = chainID

		db, dir, _, _, err := simtestutil.SetupSimulation(
			config,
			"leveldb-bApp-sim",
			"Simulation",
			simcli.FlagVerboseValue,
			simcli.FlagEnabledValue,
		)
		require.NoError(t, err, "simulation setup failed")
		t.Cleanup(func() {
			require.NoError(t, db.Close())
			require.NoError(t, os.RemoveAll(dir))
		})
	})
}

func FuzzSetDataProxyConfig(f *testing.F) {
	simcli.FlagSeedValue = time.Now().Unix()
	simcli.FlagVerboseValue = true
	simcli.FlagCommitValue = true
	simcli.FlagEnabledValue = true

	config := simcli.NewConfigFromFlags()
	config.ChainID = "foo"

	db, dir, logger, _, err := simtestutil.SetupSimulation(
		config,
		"leveldb-bApp-sim",
		"Simulation",
		simcli.FlagVerboseValue,
		simcli.FlagEnabledValue,
	)
	require.NoError(f, err, "simulation setup failed")
	f.Cleanup(func() {
		require.NoError(f, db.Close())
		require.NoError(f, os.RemoveAll(dir))
	})

	appOptions := make(simtestutil.AppOptionsMap, 0)
	appOptions[flags.FlagHome] = app.DefaultNodeHome
	appOptions[server.FlagInvCheckPeriod] = simcli.FlagPeriodValue

	bApp := app.NewApp(
		logger,
		db,
		nil,
		true,
		map[int64]bool{},
		f.TempDir(),
		0,
		appOptions,
		f.TempDir(),
		baseapp.SetChainID(config.ChainID),
	)
	require.Equal(f, app.Name, bApp.Name())
	ctx := bApp.BaseApp.NewContext(true)

	f.Fuzz(func(t *testing.T, feeAmt, feeUpdateAmt uint32, memo string) {
		privKey := secp256k1.GenPrivKey()
		pubKeyBytes := privKey.PubKey().Bytes()

		bApp.DataProxyKeeper.SetDataProxyConfig(ctx, pubKeyBytes, dataproxytypes.ProxyConfig{
			PayoutAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
			Fee: &sdk.Coin{
				Denom:  "aseda",
				Amount: cosmossdk_io_math.NewInt(int64(feeAmt)),
			},
			Memo: memo,
			FeeUpdate: &dataproxytypes.FeeUpdate{
				UpdateHeight: 0,
				NewFee: &sdk.Coin{
					Denom:  "aseda",
					Amount: cosmossdk_io_math.NewInt(int64(feeUpdateAmt)),
				},
			},
			AdminAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
		})
	})
}
