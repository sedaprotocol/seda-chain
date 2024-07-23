package types

import (
	fmt "fmt"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (msg *MsgStoreDataRequestWasm) Route() string {
	return RouterKey
}

func (msg *MsgStoreDataRequestWasm) Type() string {
	return "store-data-request-wasm"
}

func (msg *MsgStoreDataRequestWasm) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return err
	}
	if err := validateWasmSize(msg.Wasm); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}
	return nil
}

func (msg *MsgStoreOverlayWasm) Route() string {
	return RouterKey
}

func (msg *MsgStoreOverlayWasm) Type() string {
	return "store-overlay-wasm"
}

func (msg *MsgStoreOverlayWasm) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return err
	}
	if msg.WasmType != WasmTypeDataRequestExecutor && msg.WasmType != WasmTypeRelayer {
		return fmt.Errorf("overlay Wasm type must be data-request-executor or relayer")
	}
	if err := validateWasmSize(msg.Wasm); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}
	return nil
}

func (msg *MsgInstantiateAndRegisterCoreContract) Route() string {
	return RouterKey
}

func (msg *MsgInstantiateAndRegisterCoreContract) Type() string {
	return "instantiate-and-register-core-contract"
}

func (msg *MsgInstantiateAndRegisterCoreContract) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return fmt.Errorf("invalid sender: %w", err)
	}

	if msg.CodeID == 0 {
		return fmt.Errorf("code id is required")
	}

	if err := wasmtypes.ValidateLabel(msg.Label); err != nil {
		return fmt.Errorf("label is required")
	}

	if !msg.Funds.IsValid() {
		return fmt.Errorf("invalid coins")
	}

	if len(msg.Admin) != 0 {
		if _, err := sdk.AccAddressFromBech32(msg.Admin); err != nil {
			return fmt.Errorf("invalid admin: %w", err)
		}
	}
	if err := msg.Msg.ValidateBasic(); err != nil {
		return fmt.Errorf("invalid payload msg: %w", err)
	}
	if err := wasmtypes.ValidateSalt(msg.Salt); err != nil {
		return fmt.Errorf("invalid salt: %w", err)
	}
	return nil
}
