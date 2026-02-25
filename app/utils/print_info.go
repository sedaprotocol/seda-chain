package utils

import (
	"encoding/json"
	"fmt"
	"os"
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

	// Sort JSON by keys and print.
	var c any
	err = json.Unmarshal(out, &c)
	if err != nil {
		return err
	}
	js, err := json.Marshal(c)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(os.Stderr, "%s\n", js)
	return err
}
