package types

import (
	"encoding/binary"
)

// DataRequestIndex is a 56-byte index for data requests used to sort them by
// their posted gas prices and heights.
// 0                  16       24      56 (byte)
// | posted_gas_price | height | dr_id |
type DataRequestIndex []byte

func (dr DataRequest) Index() DataRequestIndex {
	// Treat gasPrice as a 128-bit unsigned integer.
	priceBytes := make([]byte, 16)
	dr.PostedGasPrice.BigInt().FillBytes(priceBytes)

	heightBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(heightBytes, dr.Height)

	drIDBytes := []byte(dr.Id)
	return append(append(priceBytes, heightBytes...), drIDBytes...)
}
