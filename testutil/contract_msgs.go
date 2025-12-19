package testutil

import (
	"encoding/hex"
	"fmt"
	"strings"
)

func CommitMsg(drID, commitment, stakerPubKey, proof string, gasUsed uint64) []byte {
	return []byte(fmt.Sprintf(`{
		"commit_data_result": {
		  "dr_id": "%s",
		  "commitment": "%s",
		  "public_key": "%s",
		  "proof": "%s",
		  "gas_used": %d
		}
	}`, drID, commitment, stakerPubKey, proof, gasUsed))
}

func RevealMsg(drID, reveal, stakerPubKey, proof string, proxyPubKeys []string, exitCode byte, drHeight, gasUsed uint64) []byte {
	quotedObjects := make([]string, 0, len(proxyPubKeys))
	for _, obj := range proxyPubKeys {
		quotedObjects = append(quotedObjects, fmt.Sprintf("%q", obj))
	}
	pks := strings.Join(quotedObjects, ",")

	return []byte(fmt.Sprintf(`{
		"reveal_data_result": {
		  "reveal_body": {
			"dr_id": "%s",
			"dr_block_height": %d,
			"exit_code": %d,
			"gas_used": %d,
			"reveal": "%s",
			"proxy_public_keys": [%s]
		  },
		  "public_key": "%s",
		  "proof": "%s",
		  "stderr": [],
		  "stdout": []
		}
	}`, drID, drHeight, exitCode, gasUsed, reveal, pks, stakerPubKey, proof))
}

func AddToAllowListMsg(stakerPubKey string) []byte {
	return []byte(fmt.Sprintf(`{
		"add_to_allowlist": {
		  "public_key": "%s"
		}
	}`, stakerPubKey))
}

func StakeMsg(stakerPubKey, proof string) []byte {
	return []byte(fmt.Sprintf(`{
		"stake": {
		  "public_key": "%s",
		  "proof": "%s",
		  "memo": "YWRkcmVzcw=="
		}
	}`, stakerPubKey, proof))
}

func PostDataRequestMsg(execProgHash, tallyProgHash []byte, requestMemo string, replicationFactor int) []byte {
	return []byte(fmt.Sprintf(`{
		"post_data_request": {
		  "posted_dr": {
			"version": "0.0.1",
			"exec_program_id": "%s",
			"exec_inputs": "ZXhlY19pbnB1dHM=",
			"exec_gas_limit": 100000000000000000,
			"tally_program_id": "%s",
			"tally_inputs": "dGFsbHlfaW5wdXRz",
			"tally_gas_limit": 300000000000000,
			"replication_factor": %d,
			"consensus_filter": "AA==",
			"gas_price": "2000",
			"memo": "%s"
		  },
		  "seda_payload": "",
		  "payback_address": "AQID"
		}
	}`, hex.EncodeToString(execProgHash), hex.EncodeToString(tallyProgHash), replicationFactor, requestMemo))
}
