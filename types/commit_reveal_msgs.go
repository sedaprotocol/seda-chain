package types

import (
	"encoding/json"
	"fmt"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// These are the JSON messages used by the overlay when executing the contract, the chain
// uses the same messages to query the contract.
type CommitDataResult struct {
	DrID      string `json:"dr_id"`
	PublicKey string `json:"public_key"`
}

type RevealDataResult struct {
	RevealBody struct {
		DrID string `json:"dr_id"`
	} `json:"reveal_body"`
	PublicKey string `json:"public_key"`
}

// ExtractCommitRevealMsgInfo checks if the message is a commit or reveal to the
// Core Contract. If it is, it returns a string of the form "drID,publicKey,isReveal"
// and true. If it is not a commit or reveal to the Core Contract, it will return
// an empty string and false.
func ExtractCommitRevealMsgInfo(coreContract string, msg sdk.Msg) (string, bool) {
	var drID, publicKey string
	var isReveal bool

	switch msg := msg.(type) {
	case *wasmtypes.MsgExecuteContract:
		if msg.Contract != coreContract {
			return "", false
		}

		contractMsg, err := unmarshalMsg(msg.Msg)
		if err != nil {
			return "", false
		}

		switch contractMsg := contractMsg.(type) {
		case CommitDataResult:
			drID = contractMsg.DrID
			publicKey = contractMsg.PublicKey
			isReveal = false
		case RevealDataResult:
			drID = contractMsg.RevealBody.DrID
			publicKey = contractMsg.PublicKey
			isReveal = true
		default:
			return "", false
		}
	default:
		return "", false
	}

	return fmt.Sprintf("%s,%s,%t", drID, publicKey, isReveal), true
}

func unmarshalMsg(msg wasmtypes.RawContractMessage) (interface{}, error) {
	var msgData struct {
		CommitDataResult *CommitDataResult `json:"commit_data_result"`
		RevealDataResult *RevealDataResult `json:"reveal_data_result"`
	}
	if err := json.Unmarshal(msg, &msgData); err != nil {
		return nil, err
	}

	if msgData.CommitDataResult != nil {
		return *msgData.CommitDataResult, nil
	}
	if msgData.RevealDataResult != nil {
		return *msgData.RevealDataResult, nil
	}
	return nil, nil
}
