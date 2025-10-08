package types

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"strconv"
)

// DataRequestIndex is a 56-byte index for data requests used to sort them by
// their posted gas prices and heights.
// 0                  16       24      56 (byte)
// | posted_gas_price | height | dr_id |
type DataRequestIndex []byte

func (i DataRequestIndex) DrID() string {
	return hex.EncodeToString(i[24:])
}

// Strings returns the string-slice representation of the data request index.
// It returns nil if the index is not valid.
// [gas_price, height, dr_id]
func (i DataRequestIndex) Strings() []string {
	if i == nil || len(i) != 56 {
		return nil
	}

	gasPrice := new(big.Int).SetBytes(i[0:16])
	height := math.MaxUint64 - binary.BigEndian.Uint64(i[16:24])
	drID := hex.EncodeToString(i[24:])

	return []string{gasPrice.String(), strconv.FormatUint(height, 10), drID}
}

func DataRequestIndexFromStrings(strings []string) (DataRequestIndex, error) {
	index := make([]byte, 56)
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
	binary.BigEndian.PutUint64(index[16:24], math.MaxUint64-height)

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
	binary.BigEndian.PutUint64(heightBytes, math.MaxUint64-uint64(dr.PostedHeight))

	drIDBytes, err := hex.DecodeString(dr.ID)
	if err != nil {
		return nil
	}
	return append(append(priceBytes, heightBytes...), drIDBytes...)
}

func (dr DataRequest) MarshalJSON() ([]byte, error) {
	type Alias DataRequest
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
