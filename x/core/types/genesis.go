package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	DefaultPaused = false
	// Testnet/Mainnet Security Group
	DefaultOwner = "seda1afk9zr2hn2jsac63h4hm60vl9z3e5u69gndzf7c99cqge3vzwjzs026662"
)

// DefaultGenesisState creates a default GenesisState object.
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		// TODO add other states
		Params: DefaultParams(),
		Paused: DefaultPaused,
		Owner:  DefaultOwner,
	}
}

// ValidateGenesis validates core genesis data.
func ValidateGenesis(state GenesisState) error {
	if state.Owner == "" {
		return sdkerrors.ErrInvalidAddress.Wrap("owner address cannot be empty")
	}

	_, err := sdk.AccAddressFromBech32(state.Owner)
	if err != nil {
		return sdkerrors.ErrInvalidAddress.Wrap("invalid owner address")
	}

	return state.Params.Validate()
}
