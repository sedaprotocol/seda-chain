package types

import (
	"bytes"
	"encoding/binary"
	"unsafe"

	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterCodec(_ *codec.LegacyAmino) {
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgUpdateParams{},
	)
	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	Amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)

const (
	b             byte   = 0
	u             uint64 = 0
	ByteLenFilter        = int(unsafe.Sizeof(b))
	ByteLenUint64        = int(unsafe.Sizeof(u))
)

func UnpackModeFilter(filterInput []byte) (string, error) {
	// Len must be at least 9 = 1 + 8
	if len(filterInput) < ByteLenFilter+ByteLenUint64 {
		return "", ErrInvalidLen.Wrapf(
			"len(filterInput): %v < %v", len(filterInput), ByteLenFilter+ByteLenUint64)
	}

	jsonPathLen := filterInput[ByteLenFilter:(ByteLenFilter + ByteLenUint64)]
	var ln uint64
	buf := bytes.NewReader(jsonPathLen)
	if err := binary.Read(buf, binary.BigEndian, &ln); err != nil {
		return "", err
	}

	data := filterInput[ByteLenFilter+ByteLenUint64:]
	// Validate the remaining length is valid.
	if len(data) != int(ln) {
		return "", ErrInvalidLen.Wrapf("want: %v Got: %v", int(ln), len(data))
	}
	return string(data), nil
}
