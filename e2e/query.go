package e2e

import (
	"encoding/json"
	"fmt"
	"net/http"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

	batchingtypes "github.com/sedaprotocol/seda-chain/x/batching/types"
	pubkeytypes "github.com/sedaprotocol/seda-chain/x/pubkey/types"
	wasmstoragetypes "github.com/sedaprotocol/seda-chain/x/wasm-storage/types"
)

func queryTx(endpoint, txHash string) error {
	resp, err := http.Get(fmt.Sprintf("%s/cosmos/tx/v1beta1/txs/%s", endpoint, txHash))
	if err != nil {
		return fmt.Errorf("failed to execute HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("tx query returned non-200 status: %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}
	txResp := result["tx_response"].(map[string]interface{})
	if v := txResp["code"]; v.(float64) != 0 {
		return fmt.Errorf("tx %s failed with status code %v", txHash, v)
	}
	return nil
}

func queryGovProposal(endpoint string, proposalID int) (govtypes.QueryProposalResponse, error) {
	var govProposalResp govtypes.QueryProposalResponse
	path := fmt.Sprintf("%s/cosmos/gov/v1/proposals/%d", endpoint, proposalID)
	body, err := httpGet(path)
	if err != nil {
		return govProposalResp, fmt.Errorf("failed to execute HTTP request: %w", err)
	}
	if err := cdc.UnmarshalJSON(body, &govProposalResp); err != nil {
		return govProposalResp, err
	}
	return govProposalResp, nil
}

func queryBatch(endpoint string, batchNumber uint64) (batchingtypes.QueryBatchResponse, error) {
	var res batchingtypes.QueryBatchResponse
	body, err := httpGet(fmt.Sprintf("%s/seda-chain/batching/batch/%s", endpoint, fmt.Sprintf("%d", batchNumber)))
	if err != nil {
		return res, err
	}
	if err = cdc.UnmarshalJSON(body, &res); err != nil {
		return res, err
	}
	return res, nil
}

func queryLatestBatch(endpoint string) (batchingtypes.QueryBatchResponse, error) {
	var res batchingtypes.QueryBatchResponse
	body, err := httpGet(fmt.Sprintf("%s/seda-chain/batching/batch/0?latest_signed=1", endpoint))
	if err != nil {
		return res, err
	}
	if err = cdc.UnmarshalJSON(body, &res); err != nil {
		return res, err
	}
	return res, nil
}

func queryPubkey(endpoint string, validatorAddress string) (pubkeytypes.QueryValidatorKeysResponse, error) {
	var res pubkeytypes.QueryValidatorKeysResponse
	body, err := httpGet(fmt.Sprintf("%s/seda-chain/pubkey/validator_keys/%s", endpoint, validatorAddress))
	if err != nil {
		return res, err
	}
	if err = cdc.UnmarshalJSON(body, &res); err != nil {
		return res, err
	}
	return res, nil
}

func queryOracleProgram(endpoint, hash string) (wasmstoragetypes.QueryOracleProgramResponse, error) {
	var res wasmstoragetypes.QueryOracleProgramResponse
	body, err := httpGet(fmt.Sprintf("%s/seda-chain/wasm-storage/oracle_program/%s", endpoint, hash))
	if err != nil {
		return res, err
	}
	if err = cdc.UnmarshalJSON(body, &res); err != nil {
		return res, err
	}
	return res, nil
}

func queryOraclePrograms(endpoint string) (wasmstoragetypes.QueryOracleProgramsResponse, error) {
	var res wasmstoragetypes.QueryOracleProgramsResponse
	body, err := httpGet(fmt.Sprintf("%s/seda-chain/wasm-storage/oracle_programs", endpoint))
	if err != nil {
		return res, err
	}
	if err = cdc.UnmarshalJSON(body, &res); err != nil {
		return res, err
	}
	return res, nil
}

func queryCoreContractRegistry(endpoint string) (wasmstoragetypes.QueryCoreContractRegistryResponse, error) {
	var res wasmstoragetypes.QueryCoreContractRegistryResponse
	body, err := httpGet(fmt.Sprintf("%s/seda-chain/wasm-storage/core_contract_registry", endpoint))
	if err != nil {
		return res, err
	}
	if err = cdc.UnmarshalJSON(body, &res); err != nil {
		return res, err
	}
	return res, nil
}
