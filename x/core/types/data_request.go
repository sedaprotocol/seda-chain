package types

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
)

// DataRequestIndex is a 56-byte index for data requests used to sort them by
// their posted gas prices and heights.
// 0                  16       24      56 (byte)
// | posted_gas_price | height | dr_id |
type DataRequestIndex []byte

func (i DataRequestIndex) DrID() string {
	return string(i[24:])
}

// Strings returns the string-slice representation of the data request index.
// It returns nil if the index is not valid.
// [gas_price, height, dr_id]
func (i DataRequestIndex) Strings() []string {
	if i == nil || len(i) != 56 {
		return nil
	}
	return []string{
		new(big.Int).SetBytes(i[0:16]).String(),
		strconv.FormatUint(binary.BigEndian.Uint64(i[16:24]), 10),
		string(i[24:]),
	}
}

func DataRequestIndexFromStrings(strings []string) (DataRequestIndex, error) {
	var index DataRequestIndex
	if len(strings) != 3 {
		return index, fmt.Errorf("required 3 components, got %d", len(strings))
	}

	// strings[0] is the posted gas price.
	postedGasPrice, ok := new(big.Int).SetString(strings[0], 10)
	if !ok {
		return index, fmt.Errorf("invalid big-int component %s", strings[0])
	}
	postedGasPriceBytes := postedGasPrice.FillBytes(make([]byte, 16))
	copy(index[0:16], postedGasPriceBytes)

	// strings[1] is the height.
	height, err := strconv.ParseUint(strings[1], 10, 64)
	if err != nil {
		return index, fmt.Errorf("invalid height component %s: %w", strings[1], err)
	}
	binary.BigEndian.PutUint64(index[16:24], height)

	// strings[2] is the dr_id.
	copy(index[24:], []byte(strings[2]))

	return index, nil
}

func (dr DataRequest) Index() DataRequestIndex {
	// Treat gasPrice as a 128-bit unsigned integer.
	priceBytes := make([]byte, 16)
	dr.PostedGasPrice.BigInt().FillBytes(priceBytes)

	heightBytes := make([]byte, 8)
	//nolint:gosec // G115: Block height is never negative.
	binary.BigEndian.PutUint64(heightBytes, uint64(dr.PostedHeight))

	drIDBytes := []byte(dr.ID) // TODO or convert hex to bytes?
	return append(append(priceBytes, heightBytes...), drIDBytes...)
}

func (dr *DataRequest) AddCommit(publicKey string, commit []byte) {
	if dr.Commits == nil {
		dr.Commits = make(map[string][]byte)
	}
	dr.Commits[publicKey] = commit
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

func (dr DataRequest) MarshalJSON() ([]byte, error) {
	type Alias DataRequest
	if dr.Commits == nil {
		dr.Commits = map[string][]byte{}
	}
	if dr.Reveals == nil {
		dr.Reveals = map[string]bool{}
	}
	if dr.ExecInputs == nil {
		dr.ExecInputs = []byte{}
	}
	if dr.TallyInputs == nil {
		dr.TallyInputs = []byte{}
	}
	if dr.ConsensusFilter == nil {
		dr.ConsensusFilter = []byte{}
	}
	if dr.Memo == nil {
		dr.Memo = []byte{}
	}
	if dr.PaybackAddress == nil {
		dr.PaybackAddress = []byte{}
	}
	if dr.SEDAPayload == nil {
		dr.SEDAPayload = []byte{}
	}
	return json.Marshal(Alias(dr))
}
