package types

import (
	"bytes"
	"encoding/binary"
)

// Sigma is a fixed-point number with 10^6 precision.
type Sigma uint64

func NewSigma(data []byte) (Sigma, error) {
	var s Sigma
	err := binary.Read(bytes.NewReader(data), binary.BigEndian, &s)
	if err != nil {
		return s, err
	}
	return s, nil
}

// WholeNumber returns Sigma's whole number part.
func (s Sigma) WholeNumber() uint64 {
	return uint64(s) / 1e6
}

// FractionalPart returns Sigma's fractional part, which is
// represented by its last six digits, as a uint64 number.
func (s Sigma) FractionalPart() uint64 {
	return uint64(s) % 1e6
}
