package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"

	"github.com/sedaprotocol/seda-chain/x/batching"
	"github.com/sedaprotocol/seda-chain/x/batching/types"
)

var validGenesisJSON = []byte(`{
      "current_batch_number": "5",
      "batches": [
        {
          "batch_number": "1",
          "block_height": "26",
          "current_data_result_root": "c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470",
          "data_result_root": "9c6b2c1b0d0b25a008e6c882cc7b415f309965c72ad2b944ac0931048ca31cd5",
          "validator_root": "328fa3aa45ae6b86c29d4a895ddf38449e3ea504de5ffa1b55389eb047b179d2",
          "batch_id": "l28f4mkRaZvM4IASbobk35htV236FzxQfhiV/P3ZHkg=",
          "proving_metadata": null
        },
        {
          "batch_number": "2",
          "block_height": "273",
          "current_data_result_root": "6bcc3a5b00afd02056ec680ee0a05306aefd4661194cf31afa31425e861bcbe2",
          "data_result_root": "54dfc95512c63c63f3d923dd613c4bb40276367a2d423bca0385754559c0e4f0",
          "validator_root": "328fa3aa45ae6b86c29d4a895ddf38449e3ea504de5ffa1b55389eb047b179d2",
          "batch_id": "wIkvbgoSQXiRsQpkO0FIsEHIVDiD/GTaRldpsTsP67Q=",
          "proving_metadata": null
        },
        {
          "batch_number": "3",
          "block_height": "294",
          "current_data_result_root": "249071c076f1fd15058f8f0791f976c9c48da39d20d73fbc3a726f205bb3af20",
          "data_result_root": "b81c09ce48830c1be6c13098d917640c096e8d76935d287013b7844358ecf2f5",
          "validator_root": "328fa3aa45ae6b86c29d4a895ddf38449e3ea504de5ffa1b55389eb047b179d2",
          "batch_id": "k7yaHK2p1OKhWFwG7nGYmabDax7InYdnjXV+b3wZeEI=",
          "proving_metadata": null
        },
        {
          "batch_number": "4",
          "block_height": "329",
          "current_data_result_root": "2726ace11e1921b15135fba087c3e99336917442c8418c15f9070112768d90ac",
          "data_result_root": "ccbca20c4a2d869f671a81a1f2d2c91fb842693ed11a4d1229c7d85bdee5a4e6",
          "validator_root": "328fa3aa45ae6b86c29d4a895ddf38449e3ea504de5ffa1b55389eb047b179d2",
          "batch_id": "J8N/92WaZ5ZGRpHyGYEWsaHodnpkPDwRMRySY9ef6wU=",
          "proving_metadata": null
        },
        {
          "batch_number": "5",
          "block_height": "342",
          "current_data_result_root": "6cd6575de411a7e69903b04c4f62dede5318fea31e7619699ee92ade44713457",
          "data_result_root": "322263c339766fc3ace8c8185c8bff5d0045389c0d8119ad6e5363bbe75bb3a2",
          "validator_root": "328fa3aa45ae6b86c29d4a895ddf38449e3ea504de5ffa1b55389eb047b179d2",
          "batch_id": "O4mEgCLBUHpU0wd1N9EWkJk7WN6wfZzjm5U+j24CqlU=",
          "proving_metadata": null
        }
      ]
    }`)

var invalidGenesisJSON = []byte(`{
      "current_batch_number": "1",
      "batches": [
        {
          "batch_number": "1",
          "block_height": "31845",
          "current_data_result_root": "f6297bfe5f9f30fb82d18b00577dc07f835711dbc26e1cfe6795445f3d514083",
          "data_result_root": "4cdcd2d3013a92929e6a5b7abfc1d14b7948c7258166d9808f40e7564f343045",
          "validator_root": "328fa3aa45ae6b86c29d4a895ddf38449e3ea504de5ffa1b55389eb047b179d2",
          "batch_id": "7O8yGG//oJuHtcRdwiiO/mVlEjCiHx56ExBGpJz3PZ4=",
          "proving_metadata": null
        }
      ]
    }`)

func TestValidateGenesis(t *testing.T) {
	cdc := moduletestutil.MakeTestEncodingConfig(batching.AppModuleBasic{}).Codec

	// Test valid JSON
	var gs types.GenesisState
	err := cdc.UnmarshalJSON(validGenesisJSON, &gs)
	require.NoError(t, err)
	err = types.ValidateGenesis(gs)
	require.NoError(t, err)

	// Test invalid JSON
	err = cdc.UnmarshalJSON(invalidGenesisJSON, &gs)
	require.NoError(t, err)
	err = types.ValidateGenesis(gs)
	require.Error(t, err)
}
