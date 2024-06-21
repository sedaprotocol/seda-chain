package keeper_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	gomock "go.uber.org/mock/gomock"

	"github.com/sedaprotocol/seda-chain/x/wasm-storage/keeper/testdata"
	"github.com/sedaprotocol/seda-wasm-vm/tallyvm"
)

// All requests use the sample tally wasm and filter type none.
var mockFetchResponse = []byte(`[{"commits":{},"dr_binary_id":"9471d36add157cd7eaa32a42b5ddd091d5d5d396bf9ad67938a4fc40209df6cf","dr_inputs":"","gas_limit":"20","gas_price":"10","height":1661661742461173125,"id":"fba5314c57e52da7d1a2245d18c670fde1cb8c237062d2a1be83f449ace0932e","memo":"","payback_address":"","replication_factor":3,"reveals":{"1b85dfb9420e6757630a0db2280fa1787ec8c1e419a6aca76dbbfe8ef6e17521":{"exit_code":0,"gas_used":"10","reveal":"Ng==","salt":"05952214b2ba3549a8d627c57d2d0dd1b0a2ce65c46e3b2f25c273464be8ba5f"},"1dae290cd880b79d21079d89aee3460cf8a7d445fb35cade70cf8aa96924441c":{"exit_code":0,"gas_used":"10","reveal":"LQ==","salt":"05952214b2ba3549a8d627c57d2d0dd1b0a2ce65c46e3b2f25c273464be8ba5f"},"421e735518ef77fc1209a9d3585cdf096669b52ea68549e2ce048d4919b4c8c0":{"exit_code":0,"gas_used":"10","reveal":"DQ==","salt":"05952214b2ba3549a8d627c57d2d0dd1b0a2ce65c46e3b2f25c273464be8ba5f"}},"seda_payload":"","tally_binary_id":"8ade60039246740faa80bf424fc29e79fe13b32087043e213e7bc36620111f6b","tally_inputs":"AAEBAQE=","version":"1.0.0"},{"commits":{},"dr_binary_id":"9471d36add157cd7eaa32a42b5ddd091d5d5d396bf9ad67938a4fc40209df6cf","dr_inputs":"","gas_limit":"20","gas_price":"10","height":9859593541233596221,"id":"d4e40f45fbf529134926acf529baeb6d4f37b5c380d7ab6b934833e7c00d725f","memo":"","payback_address":"","replication_factor":1,"reveals":{"c9a4c8f1e70a0059a88b4768a920e41c95c587b8387ea3286d8fa4ee3b68b038":{"exit_code":0,"gas_used":"10","reveal":"Yw==","salt":"f837455a930a66464f1c50586dc745a6b14ea807727c6069acac24c9558b6dbf"}},"seda_payload":"","tally_binary_id":"8ade60039246740faa80bf424fc29e79fe13b32087043e213e7bc36620111f6b","tally_inputs":"AAEBAQE=","version":"1.0.0"}]`)

func TestTallyVM(t *testing.T) {
	cases := []struct {
		name        string
		requestJSON []byte
		args        []string
		expErr      string
	}{
		{
			name:        "Three reveals",
			requestJSON: []byte(`{"commits":{},"dr_binary_id":"9471d36add157cd7eaa32a42b5ddd091d5d5d396bf9ad67938a4fc40209df6cf","dr_inputs":"","gas_limit":"20","gas_price":"10","height":1661661742461173125,"id":"fba5314c57e52da7d1a2245d18c670fde1cb8c237062d2a1be83f449ace0932e","memo":"","payback_address":"","replication_factor":3,"reveals":{"1b85dfb9420e6757630a0db2280fa1787ec8c1e419a6aca76dbbfe8ef6e17521":{"exit_code":0,"gas_used":"10","reveal":"Ng==","salt":"05952214b2ba3549a8d627c57d2d0dd1b0a2ce65c46e3b2f25c273464be8ba5f"},"1dae290cd880b79d21079d89aee3460cf8a7d445fb35cade70cf8aa96924441c":{"exit_code":0,"gas_used":"10","reveal":"LQ==","salt":"05952214b2ba3549a8d627c57d2d0dd1b0a2ce65c46e3b2f25c273464be8ba5f"},"421e735518ef77fc1209a9d3585cdf096669b52ea68549e2ce048d4919b4c8c0":{"exit_code":0,"gas_used":"10","reveal":"DQ==","salt":"05952214b2ba3549a8d627c57d2d0dd1b0a2ce65c46e3b2f25c273464be8ba5f"}},"seda_payload":"","tally_binary_id":"8ade60039246740faa80bf424fc29e79fe13b32087043e213e7bc36620111f6b","tally_inputs":"AAEBAQE=","version":"1.0.0"}`),
			args:        []string{`\x00\x01\x01\x01\x01`, "[{\"reveal\":[54],\"salt\":[211,159,121,219,109,120,111,102,218,223,158,61,107,199,122,219,183,57,237,221,157,209,215,117,111,70,182,113,238,185,115,142,158,221,189,159,219,151,54,239,126,58,225,183,188,109,174,95],\"exit_code\":0,\"gas_used\":\"10\"},{\"reveal\":[45],\"salt\":[211,159,121,219,109,120,111,102,218,223,158,61,107,199,122,219,183,57,237,221,157,209,215,117,111,70,182,113,238,185,115,142,158,221,189,159,219,151,54,239,126,58,225,183,188,109,174,95],\"exit_code\":0,\"gas_used\":\"10\"},{\"reveal\":[13],\"salt\":[211,159,121,219,109,120,111,102,218,223,158,61,107,199,122,219,183,57,237,221,157,209,215,117,111,70,182,113,238,185,115,142,158,221,189,159,219,151,54,239,126,58,225,183,188,109,174,95],\"exit_code\":0,\"gas_used\":\"10\"}]", "[0,0,0]"},
			expErr:      "",
		},
		{
			name:        "One less outlier provided",
			requestJSON: []byte(`{"commits":{},"dr_binary_id":"9471d36add157cd7eaa32a42b5ddd091d5d5d396bf9ad67938a4fc40209df6cf","dr_inputs":"","gas_limit":"20","gas_price":"10","height":1661661742461173125,"id":"fba5314c57e52da7d1a2245d18c670fde1cb8c237062d2a1be83f449ace0932e","memo":"","payback_address":"","replication_factor":3,"reveals":{"1b85dfb9420e6757630a0db2280fa1787ec8c1e419a6aca76dbbfe8ef6e17521":{"exit_code":0,"gas_used":"10","reveal":"Ng==","salt":"05952214b2ba3549a8d627c57d2d0dd1b0a2ce65c46e3b2f25c273464be8ba5f"},"1dae290cd880b79d21079d89aee3460cf8a7d445fb35cade70cf8aa96924441c":{"exit_code":0,"gas_used":"10","reveal":"LQ==","salt":"05952214b2ba3549a8d627c57d2d0dd1b0a2ce65c46e3b2f25c273464be8ba5f"},"421e735518ef77fc1209a9d3585cdf096669b52ea68549e2ce048d4919b4c8c0":{"exit_code":0,"gas_used":"10","reveal":"DQ==","salt":"05952214b2ba3549a8d627c57d2d0dd1b0a2ce65c46e3b2f25c273464be8ba5f"}},"seda_payload":"","tally_binary_id":"8ade60039246740faa80bf424fc29e79fe13b32087043e213e7bc36620111f6b","tally_inputs":"AAEBAQE=","version":"1.0.0"}`),
			args:        []string{`\x00\x01\x01\x01\x01`, "[{\"reveal\":[54],\"salt\":[211,159,121,219,109,120,111,102,218,223,158,61,107,199,122,219,183,57,237,221,157,209,215,117,111,70,182,113,238,185,115,142,158,221,189,159,219,151,54,239,126,58,225,183,188,109,174,95],\"exit_code\":0,\"gas_used\":\"10\"},{\"reveal\":[45],\"salt\":[211,159,121,219,109,120,111,102,218,223,158,61,107,199,122,219,183,57,237,221,157,209,215,117,111,70,182,113,238,185,115,142,158,221,189,159,219,151,54,239,126,58,225,183,188,109,174,95],\"exit_code\":0,\"gas_used\":\"10\"},{\"reveal\":[13],\"salt\":[211,159,121,219,109,120,111,102,218,223,158,61,107,199,122,219,183,57,237,221,157,209,215,117,111,70,182,113,238,185,115,142,158,221,189,159,219,151,54,239,126,58,225,183,188,109,174,95],\"exit_code\":0,\"gas_used\":\"10\"}]", "[0,0]"},
			expErr:      "abort: Number of reveals (3) does not equal number of consensus reports (2). in assembly/tally.ts(51:7)",
		},
		{
			name:        "One reveal",
			requestJSON: []byte(`{"commits":{},"dr_binary_id":"9471d36add157cd7eaa32a42b5ddd091d5d5d396bf9ad67938a4fc40209df6cf","dr_inputs":"","gas_limit":"20","gas_price":"10","height":9859593541233596221,"id":"d4e40f45fbf529134926acf529baeb6d4f37b5c380d7ab6b934833e7c00d725f","memo":"","payback_address":"","replication_factor":1,"reveals":{"c9a4c8f1e70a0059a88b4768a920e41c95c587b8387ea3286d8fa4ee3b68b038":{"exit_code":0,"gas_used":"10","reveal":"Yw==","salt":"f837455a930a66464f1c50586dc745a6b14ea807727c6069acac24c9558b6dbf"}},"seda_payload":"","tally_binary_id":"8ade60039246740faa80bf424fc29e79fe13b32087043e213e7bc36620111f6b","tally_inputs":"AAEBAQE=","version":"1.0.0"}`),
			args:        []string{"\x00\x01\x01\x01\x01", "[{\"reveal\":[99],\"salt\":[127,205,251,227,158,90,247,125,26,235,174,58,225,253,92,231,78,124,233,215,59,227,150,186,111,94,30,107,205,59,239,110,220,235,78,189,105,198,156,219,135,61,231,159,27,233,214,223],\"exit_code\":0,\"gas_used\":\"10\"}]", "[0]"},
			expErr:      "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := tallyvm.ExecuteTallyVm(testdata.SampleTallyWasm(), tc.args, map[string]string{
				"VM_MODE":   "tally",
				"CONSENSUS": "true",
			})
			fmt.Println(result)

			if tc.expErr != "" {
				require.Contains(t, result.Stderr[0], tc.expErr)
			}

			// var vmResDbg []VMResult
			// err = json.Unmarshal(result.Result, &vmResDbg)
			// if err != nil {
			// 	panic(err)
			// }
			// fmt.Printf("%+v\n", vmResDbg)
		})
	}
}

// TestExecuteTally tests ExecuteTally using mock fetch data up to the
// point where VM execution results are posted to the contract.
func TestExecuteTally(t *testing.T) {
	f := initFixture(t)
	ctx := f.Context()

	// Mock response for fetching tally-ready data requests.
	f.mockViewKeeper.EXPECT().QuerySmart(gomock.Any(), gomock.Any(), gomock.Any()).Return(mockFetchResponse, nil)

	// Contract should return not found in response to post data result
	// since the fetch data was mocked.
	err := f.wasmStorageKeeper.ExecuteTally(ctx)
	require.Contains(t, err.Error(), "seda_contract::msgs::data_requests::types::StatusValue; key:")
	require.Contains(t, err.Error(), "not found")
}
