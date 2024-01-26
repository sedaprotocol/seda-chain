package utils

import (
	"encoding/json"
	"fmt"
	"os"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type PrintInfo struct {
	Moniker string `json:"moniker" yaml:"moniker"`
	ChainID string `json:"chain_id" yaml:"chain_id"`
	NodeID  string `json:"node_id" yaml:"node_id"`
	Seeds   string `json:"seeds" yaml:"seeds"`
}

func NewPrintInfo(moniker, chainID, nodeID, seeds string) PrintInfo {
	return PrintInfo{
		Moniker: moniker,
		ChainID: chainID,
		NodeID:  nodeID,
		Seeds:   seeds,
	}
}

func DisplayInfo(info PrintInfo) error {
	out, err := json.MarshalIndent(info, "", " ")
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(os.Stderr, "%s\n", sdk.MustSortJSON(out))
	return err
}
