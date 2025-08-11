package types

import (
	"encoding/binary"
)

// DataRequestIndex is a 56-byte index for data requests used to sort them by
// their posted gas prices and heights.
// 0                  16       24      56 (byte)
// | posted_gas_price | height | dr_id |
type DataRequestIndex []byte

func (i DataRequestIndex) DrID() string {
	return string(i[24:])
}

func (dr DataRequest) Index() DataRequestIndex {
	// Treat gasPrice as a 128-bit unsigned integer.
	priceBytes := make([]byte, 16)
	dr.PostedGasPrice.BigInt().FillBytes(priceBytes)

	heightBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(heightBytes, dr.Height)

	drIDBytes := []byte(dr.Id) // TODO or convert hex to bytes?
	return append(append(priceBytes, heightBytes...), drIDBytes...)
}

func (dr *DataRequest) AddCommit(publicKey string, commitment []byte) {
	if dr.Commits == nil {
		dr.Commits = make(map[string][]byte)
	}
	dr.Commits[publicKey] = commitment
}

func (dr DataRequest) GetCommit(publicKey string) ([]byte, bool) {
	if dr.Commits == nil {
		return nil, false
	}
	commit, exists := dr.Commits[publicKey]
	return commit, exists
}

// MarkAsRevealed adds the given public key to the data request's reveals map
// and returns the count of reveals.
func (dr *DataRequest) MarkAsRevealed(publicKey string) int {
	if dr.Reveals == nil {
		dr.Reveals = make(map[string]bool)
	}
	dr.Reveals[publicKey] = true
	return len(dr.Reveals)
}

func (dr DataRequest) HasRevealed(publicKey string) bool {
	if dr.Reveals == nil {
		return false
	}
	_, exists := dr.Reveals[publicKey]
	return exists
}
