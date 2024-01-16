package types

import "github.com/strangelove-ventures/interchaintest/v8/ibc"

type GetCountObj struct {
	Count int64 `json:"count"`
}

type GetCountResponse struct {
	Data *GetCountObj `json:"data"`
}

type RelayerConfig struct {
	Type    ibc.RelayerImplementation
	Name    string
	Image   string
	Version string
}
