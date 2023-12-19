package utils

import (
	"encoding/json"
	"fmt"
	"os"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type printInfo struct {
	Moniker string `json:"moniker" yaml:"moniker"`
	ChainID string `json:"chain_id" yaml:"chain_id"`
	NodeID  string `json:"node_id" yaml:"node_id"`
	Seeds   string `json:"seeds" yaml:"seeds"`
}

func NewPrintInfo(moniker, chainID, nodeID, seeds string) printInfo {
	return printInfo{
		Moniker: moniker,
		ChainID: chainID,
		NodeID:  nodeID,
		Seeds:   seeds,
	}
}

func DisplayInfo(info printInfo) error {
	out, err := json.MarshalIndent(info, "", " ")
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(os.Stderr, "%s\n", sdk.MustSortJSON(out))
	return err
}
