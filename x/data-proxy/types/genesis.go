package types

import "fmt"

func NewGenesisState(params Params, proxyConfigs []DataProxyConfig, feeUpdates []FeeUpdateQueueRecord) GenesisState {
	return GenesisState{
		Params:           params,
		DataProxyConfigs: proxyConfigs,
		FeeUpdateQueue:   feeUpdates,
	}
}

func DefaultGenesisState() *GenesisState {
	state := NewGenesisState(DefaultParams(), []DataProxyConfig{}, []FeeUpdateQueueRecord{})
	return &state
}

func ValidateGenesis(data GenesisState) error {
	for _, proxyConfig := range data.DataProxyConfigs {
		if len(proxyConfig.DataProxyPubkey) == 0 {
			return fmt.Errorf("empty public key in proxy configs")
		}

		if err := proxyConfig.Config.Validate(); err != nil {
			return err
		}

		if err := proxyConfig.Config.Fee.Validate(); err != nil {
			return err
		}

		if proxyConfig.Config.AdminAddress == "" {
			return fmt.Errorf("empty admin address")
		}
		if proxyConfig.Config.PayoutAddress == "" {
			return fmt.Errorf("empty payout address")
		}
	}

	for _, feeUpdate := range data.FeeUpdateQueue {
		if len(feeUpdate.DataProxyPubkey) == 0 {
			return fmt.Errorf("empty public key in fee updates")
		}
	}

	return data.Params.Validate()
}
