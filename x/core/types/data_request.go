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
// 0                 16                24      56 (byte)
// | posted_gas_price | inverted_height | dr_id |
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
	if len(strings[2]) != 64 {
		return index, fmt.Errorf("invalid data request ID component %s: must be 64 characters long", strings[2])
	}

	// strings[0] is the posted gas price.
	postedGasPrice, ok := new(big.Int).SetString(strings[0], 10)
	if !ok {
		return index, fmt.Errorf("invalid gas price component %s", strings[0])
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
	drIDBytes, err := hex.DecodeString(strings[2])
	if err != nil {
		return index, fmt.Errorf("invalid data request ID component %s: %w", strings[2], err)
	}
	copy(index[24:], drIDBytes)

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

// MarshalJSON flattens the JSON serialization of DataRequestResponse and
// initializes nil fields to avoid their omission.
func (dr DataRequestResponse) MarshalJSON() ([]byte, error) {
	flattened := struct {
		DataRequest
		Commits map[string][]byte      `json:"commits"`
		Reveals map[string]*RevealBody `json:"reveals"`
	}{
		DataRequest: dr.DataRequest,
		Commits:     dr.Commits,
		Reveals:     dr.Reveals,
	}

	if flattened.ExecInputs == nil {
		flattened.ExecInputs = []byte{}
	}
	if flattened.TallyInputs == nil {
		flattened.TallyInputs = []byte{}
	}
	if flattened.ConsensusFilter == nil {
		flattened.ConsensusFilter = []byte{}
	}
	if flattened.Memo == nil {
		flattened.Memo = []byte{}
	}
	if flattened.PaybackAddress == nil {
		flattened.PaybackAddress = []byte{}
	}
	if flattened.SEDAPayload == nil {
		flattened.SEDAPayload = []byte{}
	}

	if flattened.Commits == nil {
		flattened.Commits = map[string][]byte{}
	}
	if flattened.Reveals == nil {
		flattened.Reveals = map[string]*RevealBody{}
	}

	bz, err := json.Marshal(flattened)
	if err != nil {
		return nil, err
	}
	return bz, nil
}
