package keeper_test

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sedaprotocol/seda-wasm-vm/tallyvm/v2"

	"github.com/sedaprotocol/seda-chain/x/tally/keeper"
	"github.com/sedaprotocol/seda-chain/x/tally/keeper/testdata"
	"github.com/sedaprotocol/seda-chain/x/tally/types"
)

func TestEndBlock(t *testing.T) {
	f := initFixture(t)

	tests := []struct {
		name              string
		memo              string
		replicationFactor int
		numCommits        int
		numReveals        int
		timeout           bool
		expExitCode       uint32
	}{
		{
			name:              "full single commit-reveal",
			memo:              "YzJWamRYSmxaR0YwWVE9PQ==",
			replicationFactor: 1,
			numCommits:        1,
			numReveals:        1,
			timeout:           false,
			expExitCode:       0,
		},
		{
			name:              "full 5 commit-reveals",
			memo:              "ZnVsbCA1IGNvbW1pdC1yZXZlYWxz",
			replicationFactor: 5,
			numCommits:        5,
			numReveals:        5,
			timeout:           false,
			expExitCode:       0,
		},
		{
			name:              "commit timeout",
			memo:              "Y29tbWl0IHRpbWVvdXQ=",
			replicationFactor: 2,
			numCommits:        0,
			numReveals:        0,
			timeout:           true,
			expExitCode:       keeper.TallyExitCodeNotEnoughCommits,
		},
		{
			name:              "commit timeout with 1 commit",
			memo:              "Y29tbWl0IHRpbWVvdXQgd2l0aCAxIGNvbW1pdA==",
			replicationFactor: 2,
			numCommits:        1,
			numReveals:        0,
			timeout:           true,
			expExitCode:       keeper.TallyExitCodeNotEnoughCommits,
		},
		{
			name:              "commit timeout with 2 commits",
			memo:              "Y29tbWl0IHRpbWVvdXQgd2l0aCAyIGNvbW1pdHM=",
			replicationFactor: 2,
			numCommits:        1,
			numReveals:        0,
			timeout:           true,
			expExitCode:       keeper.TallyExitCodeNotEnoughCommits,
		},
		{
			name:              "reveal timeout",
			memo:              "cmV2ZWFsIHRpbWVvdXQ=",
			replicationFactor: 2,
			numCommits:        2,
			numReveals:        0,
			timeout:           true,
			expExitCode:       keeper.TallyExitCodeNotEnoughReveals,
		},
		{
			name:              "reveal timeout with 2 reveals",
			memo:              "cmV2ZWFsIHRpbWVvdXQgd2l0aCAyIHJldmVhbHM=",
			replicationFactor: 3,
			numCommits:        3,
			numReveals:        2,
			timeout:           true,
			expExitCode:       keeper.TallyExitCodeNotEnoughReveals,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			drID := f.commitRevealDataRequest(t, tt.memo, tt.replicationFactor, tt.numCommits, tt.numReveals, tt.timeout)

			err := f.tallyKeeper.EndBlock(f.Context())
			require.NoError(t, err)

			dataResult, err := f.batchingKeeper.GetLatestDataResult(f.Context(), drID)
			require.NoError(t, err)
			require.Equal(t, tt.expExitCode, dataResult.ExitCode)

			dataResults, err := f.batchingKeeper.GetDataResults(f.Context(), false)
			require.NoError(t, err)
			require.Contains(t, dataResults, *dataResult)
		})
	}
}

// TestTallyVM tests tally VM using a sample tally wasm that performs
// preliminary checks on the given reveal data.
func TestTallyVM(t *testing.T) {
	cases := []struct {
		name        string
		requestJSON []byte
		args        []string
		expErr      string
	}{
		{
			name: "Three reveals",
			requestJSON: []byte(`{
				"commits":{},
				"exec_program_id":"9471d36add157cd7eaa32a42b5ddd091d5d5d396bf9ad67938a4fc40209df6cf",
				"exec_inputs":"",
				"exec_gas_limit":5000000000,
				"gas_price":"10",
				"height":1661661742461173125,
				"id":"fba5314c57e52da7d1a2245d18c670fde1cb8c237062d2a1be83f449ace0932e",
				"memo":"",
				"payback_address":"",
				"replication_factor":3,
				"reveals":{
				   "1b85dfb9420e6757630a0db2280fa1787ec8c1e419a6aca76dbbfe8ef6e17521":{
					  "exit_code":0,
					  "gas_used":10,
					  "reveal":"Ng==",
					  "salt":"05952214b2ba3549a8d627c57d2d0dd1b0a2ce65c46e3b2f25c273464be8ba5f"
				   },
				   "1dae290cd880b79d21079d89aee3460cf8a7d445fb35cade70cf8aa96924441c":{
					  "exit_code":0,
					  "gas_used":10,
					  "reveal":"LQ==",
					  "salt":"05952214b2ba3549a8d627c57d2d0dd1b0a2ce65c46e3b2f25c273464be8ba5f"
				   },
				   "421e735518ef77fc1209a9d3585cdf096669b52ea68549e2ce048d4919b4c8c0":{
					  "exit_code":0,
					  "gas_used":10,
					  "reveal":"DQ==",
					  "salt":"05952214b2ba3549a8d627c57d2d0dd1b0a2ce65c46e3b2f25c273464be8ba5f"
				   }
				},
				"seda_payload":"",
				"tally_program_id":"8ade60039246740faa80bf424fc29e79fe13b32087043e213e7bc36620111f6b",
				"tally_inputs":"AAEBAQE=",
				"tally_gas_limit":50000000000000,
				"version":"1.0.0"
			 }`),
			args:   []string{"6d792d74616c6c792d696e70757473", "[{\"reveal\":[54],\"salt\":[211,159,121,219,109,120,111,102,218,223,158,61,107,199,122,219,183,57,237,221,157,209,215,117,111,70,182,113,238,185,115,142,158,221,189,159,219,151,54,239,126,58,225,183,188,109,174,95],\"exit_code\":0,\"gas_used\":\"10\"},{\"reveal\":[45],\"salt\":[211,159,121,219,109,120,111,102,218,223,158,61,107,199,122,219,183,57,237,221,157,209,215,117,111,70,182,113,238,185,115,142,158,221,189,159,219,151,54,239,126,58,225,183,188,109,174,95],\"exit_code\":0,\"gas_used\":\"10\"},{\"reveal\":[13],\"salt\":[211,159,121,219,109,120,111,102,218,223,158,61,107,199,122,219,183,57,237,221,157,209,215,117,111,70,182,113,238,185,115,142,158,221,189,159,219,151,54,239,126,58,225,183,188,109,174,95],\"exit_code\":0,\"gas_used\":\"10\"}]", "[0,0,0]"},
			expErr: "",
		},
		{
			name: "One less outlier provided",
			requestJSON: []byte(`{
				"commits":{},
				"exec_program_id":"9471d36add157cd7eaa32a42b5ddd091d5d5d396bf9ad67938a4fc40209df6cf",
				"exec_inputs":"",
				"exec_gas_limit":5000000000,
				"gas_price":"10",
				"height":1661661742461173125,
				"id":"fba5314c57e52da7d1a2245d18c670fde1cb8c237062d2a1be83f449ace0932e",
				"memo":"",
				"payback_address":"",
				"replication_factor":3,
				"reveals":{
				   "1b85dfb9420e6757630a0db2280fa1787ec8c1e419a6aca76dbbfe8ef6e17521":{
					  "exit_code":0,
					  "gas_used":10,
					  "reveal":"Ng==",
					  "salt":"05952214b2ba3549a8d627c57d2d0dd1b0a2ce65c46e3b2f25c273464be8ba5f"
				   },
				   "1dae290cd880b79d21079d89aee3460cf8a7d445fb35cade70cf8aa96924441c":{
					  "exit_code":0,
					  "gas_used":10,
					  "reveal":"LQ==",
					  "salt":"05952214b2ba3549a8d627c57d2d0dd1b0a2ce65c46e3b2f25c273464be8ba5f"
				   },
				   "421e735518ef77fc1209a9d3585cdf096669b52ea68549e2ce048d4919b4c8c0":{
					  "exit_code":0,
					  "gas_used":10,
					  "reveal":"DQ==",
					  "salt":"05952214b2ba3549a8d627c57d2d0dd1b0a2ce65c46e3b2f25c273464be8ba5f"
				   }
				},
				"seda_payload":"",
				"tally_program_id":"8ade60039246740faa80bf424fc29e79fe13b32087043e213e7bc36620111f6b",
				"tally_inputs":"AAEBAQE=",
				"tally_gas_limit":50000000000000,
				"version":"1.0.0"
			 }`),
			args:   []string{"6d792d74616c6c792d696e70757473", "[{\"reveal\":[54],\"salt\":[211,159,121,219,109,120,111,102,218,223,158,61,107,199,122,219,183,57,237,221,157,209,215,117,111,70,182,113,238,185,115,142,158,221,189,159,219,151,54,239,126,58,225,183,188,109,174,95],\"exit_code\":0,\"gas_used\":\"10\"},{\"reveal\":[45],\"salt\":[211,159,121,219,109,120,111,102,218,223,158,61,107,199,122,219,183,57,237,221,157,209,215,117,111,70,182,113,238,185,115,142,158,221,189,159,219,151,54,239,126,58,225,183,188,109,174,95],\"exit_code\":0,\"gas_used\":\"10\"},{\"reveal\":[13],\"salt\":[211,159,121,219,109,120,111,102,218,223,158,61,107,199,122,219,183,57,237,221,157,209,215,117,111,70,182,113,238,185,115,142,158,221,189,159,219,151,54,239,126,58,225,183,188,109,174,95],\"exit_code\":0,\"gas_used\":\"10\"}]", "[0,0]"},
			expErr: "abort: Number of reveals (3) does not equal number of consensus reports (2)",
		},
		{
			name: "One reveal",
			requestJSON: []byte(`{
				"commits":{},
				"exec_program_id":"9471d36add157cd7eaa32a42b5ddd091d5d5d396bf9ad67938a4fc40209df6cf",
				"exec_inputs":"",
				"exec_gas_limit":5000000000,
				"gas_price":"10",
				"height":9859593541233596221,
				"id":"d4e40f45fbf529134926acf529baeb6d4f37b5c380d7ab6b934833e7c00d725f",
				"memo":"",
				"payback_address":"",
				"replication_factor":1,
				"reveals":{
				   "c9a4c8f1e70a0059a88b4768a920e41c95c587b8387ea3286d8fa4ee3b68b038":{
					  "exit_code":0,
					  "gas_used":10,
					  "reveal":"Yw==",
					  "salt":"f837455a930a66464f1c50586dc745a6b14ea807727c6069acac24c9558b6dbf"
				   }
				},
				"seda_payload":"",
				"tally_program_id":"8ade60039246740faa80bf424fc29e79fe13b32087043e213e7bc36620111f6b",
				"tally_inputs":"AAEBAQE=",
				"tally_gas_limit":50000000000000,
				"version":"1.0.0"
			 }`),
			args:   []string{"6d792d74616c6c792d696e70757473", "[{\"reveal\":[99],\"salt\":[127,205,251,227,158,90,247,125,26,235,174,58,225,253,92,231,78,124,233,215,59,227,150,186,111,94,30,107,205,59,239,110,220,235,78,189,105,198,156,219,135,61,231,159,27,233,214,223],\"exit_code\":0,\"gas_used\":\"10\"}]", "[0]"},
			expErr: "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var req types.Request
			err := json.Unmarshal(tc.requestJSON, &req)
			require.NoError(t, err)

			result := tallyvm.ExecuteTallyVm(testdata.SampleTallyWasm(), tc.args, map[string]string{
				"VM_MODE":               "tally",
				"CONSENSUS":             fmt.Sprintf("%v", true),
				"BLOCK_HEIGHT":          fmt.Sprintf("%d", 1),
				"DR_ID":                 req.ID,
				"EXEC_PROGRAM_ID":       req.ExecProgramID,
				"EXEC_INPUTS":           req.ExecInputs,
				"EXEC_GAS_LIMIT":        fmt.Sprintf("%v", req.ExecGasLimit),
				"TALLY_INPUTS":          req.TallyInputs,
				"TALLY_PROGRAM_ID":      req.TallyProgramID,
				"DR_REPLICATION_FACTOR": fmt.Sprintf("%v", req.ReplicationFactor),
				"DR_GAS_PRICE":          req.GasPrice,
				"DR_TALLY_GAS_LIMIT":    fmt.Sprintf("%v", req.TallyGasLimit),
				"DR_MEMO":               req.Memo,
				"DR_PAYBACK_ADDRESS":    req.PaybackAddress,
			})

			if tc.expErr != "" {
				require.Contains(t, result.Stderr[0], tc.expErr)
			} else {
				require.Equal(t, 0, len(result.Stderr))

				bz, err := hex.DecodeString(tc.args[0])
				require.NoError(t, err)
				require.Contains(t, string(result.Result), string(bz))
			}
		})
	}
}

// TestTallyVM_EnvVars tests passing environment variables to tally VM.
func TestTallyVM_EnvVars(t *testing.T) {
	cases := []struct {
		name        string
		requestJSON []byte
		args        []string
		expResult   string
		expErr      string
	}{
		{
			name: "Test passing all environment variables",
			requestJSON: []byte(`{
				"commits":{},
				"exec_program_id":"9471d36add157cd7eaa32a42b5ddd091d5d5d396bf9ad67938a4fc40209df6cf",
				"exec_inputs":"",
				"exec_gas_limit":5000000000,
				"gas_price":"10",
				"height":1661661742461173200,
				"id":"fba5314c57e52da7d1a2245d18c670fde1cb8c237062d2a1be83f449ace0932e",
				"memo":"mock_data_request_num_one",
				"payback_address":"YrzimoSJXwpA7ju71AkhkirkDCU=",
				"consensus_filter":"AQAAAAAAAAALcmVzdWx0LnRleHQ=",
				"replication_factor":3,
				"reveals":{},
				"seda_payload":"",
				"tally_program_id":"5f3b31bff28c64a143119ee6389d62e38767672daace9c36db54fa2d18e9f391",
				"tally_inputs":"AAEBAQE=",
				"tally_gas_limit":50000000000000,
				"version":"1.0.0"
			}`),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var req types.Request
			err := json.Unmarshal(tc.requestJSON, &req)
			require.NoError(t, err)

			envs := map[string]string{
				"VM_MODE":               "tally",
				"CONSENSUS":             fmt.Sprintf("%v", true),
				"BLOCK_HEIGHT":          fmt.Sprintf("%d", 1),
				"DR_ID":                 req.ID,
				"EXEC_PROGRAM_ID":       req.ExecProgramID,
				"EXEC_INPUTS":           req.ExecInputs,
				"EXEC_GAS_LIMIT":        fmt.Sprintf("%v", req.ExecGasLimit),
				"TALLY_INPUTS":          req.TallyInputs,
				"TALLY_PROGRAM_ID":      req.TallyProgramID,
				"DR_REPLICATION_FACTOR": fmt.Sprintf("%v", req.ReplicationFactor),
				"DR_GAS_PRICE":          req.GasPrice,
				"DR_TALLY_GAS_LIMIT":    fmt.Sprintf("%v", req.TallyGasLimit),
				"DR_MEMO":               req.Memo,
				"DR_PAYBACK_ADDRESS":    req.PaybackAddress,
			}

			result := tallyvm.ExecuteTallyVm(testdata.SampleTallyWasm2(), tc.args, envs)

			require.Equal(t, 0, len(result.Stderr))
			for key := range envs {
				require.Contains(t, string(result.Result), fmt.Sprintf("%s=%s", key, envs[key]))
			}
		})
	}
}
