package e2e

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

	wasmstoragetypes "github.com/sedaprotocol/seda-chain/x/wasm-storage/types"
)

func query(endpoint, path string, v interface{}) error {
	resp, err := http.Get(fmt.Sprintf("%s/%s", endpoint, path))
	if err != nil {
		return fmt.Errorf("failed to execute HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("query returned non-200 status: %d, body: %s", resp.StatusCode, body)
	}

	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	return nil
}

func queryTx(endpoint, txHash string) error {
	var result map[string]interface{}
	err := query(endpoint, fmt.Sprintf("cosmos/tx/v1beta1/txs/%s", txHash), &result)
	if err != nil {
		return err
	}

	txResp := result["tx_response"].(map[string]interface{})
	if v := txResp["code"]; v.(float64) != 0 {
		return fmt.Errorf("tx %s failed with status code %v", txHash, v)
	}

	return nil
}

func queryGovProposal(endpoint string, proposalID int) (govtypes.QueryProposalResponse, error) {
	var govProposalResp govtypes.QueryProposalResponse
	err := query(endpoint, fmt.Sprintf("cosmos/gov/v1/proposals/%d", proposalID), &govProposalResp)
	return govProposalResp, err
}

func queryDataRequestWasm(endpoint string, drHash string) (wasmstoragetypes.QueryDataRequestWasmResponse, error) {
	var res wasmstoragetypes.QueryDataRequestWasmResponse
	err := query(endpoint, fmt.Sprintf("seda-chain/wasm-storage/data_request_wasm/%s", drHash), &res)
	return res, err
}

func queryOverlayWasm(endpoint string, hash string) (wasmstoragetypes.QueryOverlayWasmResponse, error) {
	var res wasmstoragetypes.QueryOverlayWasmResponse
	err := query(endpoint, fmt.Sprintf("seda-chain/wasm-storage/overlay_wasm/%s", hash), &res)
	return res, err
}

func queryDataRequestWasms(endpoint string) (wasmstoragetypes.QueryDataRequestWasmsResponse, error) {
	var res wasmstoragetypes.QueryDataRequestWasmsResponse
	err := query(endpoint, "seda-chain/wasm-storage/data_request_wasms", &res)
	return res, err
}

func queryOverlayWasms(endpoint string) (wasmstoragetypes.QueryOverlayWasmsResponse, error) {
	var res wasmstoragetypes.QueryOverlayWasmsResponse
	err := query(endpoint, "seda-chain/wasm-storage/overlay_wasms", &res)
	return res, err
}

func queryProxyContractRegistry(endpoint string) (wasmstoragetypes.QueryProxyContractRegistryResponse, error) {
	var res wasmstoragetypes.QueryProxyContractRegistryResponse
	err := query(endpoint, "seda-chain/wasm-storage/proxy_contract_registry", &res)
	return res, err
}
