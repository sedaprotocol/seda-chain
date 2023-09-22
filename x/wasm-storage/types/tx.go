package types

import (
	fmt "fmt"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (msg MsgStoreDataRequestWasm) Route() string {
	return RouterKey
}

func (msg MsgStoreDataRequestWasm) Type() string {
	return "store-data-request-wasm"
}

func (msg MsgStoreDataRequestWasm) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return err
	}
	if msg.WasmType != WasmTypeDataRequest && msg.WasmType != WasmTypeTally {
		return fmt.Errorf("Data Request Wasm type must be data-request or tally")
	}
	if err := validateWasmCode(msg.Wasm); err != nil {
		return fmt.Errorf("invalid request: code bytes %s", err.Error())
	}
	return nil
}

func (msg MsgStoreDataRequestWasm) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgStoreDataRequestWasm) GetSigners() []sdk.AccAddress {
	senderAddr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil { // should never happen as ValidateBasic() rejects invalid addresses
		panic(err.Error())
	}
	return []sdk.AccAddress{senderAddr}
}

func (msg MsgStoreOverlayWasm) Route() string {
	return RouterKey
}

func (msg MsgStoreOverlayWasm) Type() string {
	return "store-overlay-wasm"
}

func (msg MsgStoreOverlayWasm) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return err
	}
	if msg.WasmType != WasmTypeDataRequestExecutor && msg.WasmType != WasmTypeRelayer {
		return fmt.Errorf("Overlay Wasm type must be data-request-executor or relayer")
	}
	if err := validateWasmCode(msg.Wasm); err != nil {
		return fmt.Errorf("invalid request: code bytes %s", err.Error())
	}
	return nil
}

func (msg MsgStoreOverlayWasm) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgStoreOverlayWasm) GetSigners() []sdk.AccAddress {
	senderAddr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil { // should never happen as ValidateBasic() rejects invalid addresses
		panic(err.Error())
	}
	return []sdk.AccAddress{senderAddr}
}

func (msg MsgInstantiateAndRegisterProxyContract) Route() string {
	return RouterKey
}

func (msg MsgInstantiateAndRegisterProxyContract) Type() string {
	return "instantiate-and-register-proxy-contract"
}

func (msg MsgInstantiateAndRegisterProxyContract) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return fmt.Errorf("invalid sender: %s", err)
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
			return fmt.Errorf("invalid admin: %s", err)
		}
	}
	if err := msg.Msg.ValidateBasic(); err != nil {
		return fmt.Errorf("invalid payload msg: %s", err)
	}
	if err := wasmtypes.ValidateSalt(msg.Salt); err != nil {
		return fmt.Errorf("invalid salt: %s", err)
	}
	return nil
}

func (msg MsgInstantiateAndRegisterProxyContract) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgInstantiateAndRegisterProxyContract) GetSigners() []sdk.AccAddress {
	senderAddr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil { // should never happen as ValidateBasic() rejects invalid addresses
		panic(err.Error())
	}
	return []sdk.AccAddress{senderAddr}
}
