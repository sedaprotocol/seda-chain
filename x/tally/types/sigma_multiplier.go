package types

import (
	"bytes"
	"encoding/binary"
	"math/big"
)

// SigmaMultiplier is a 10^6 precision fixed-point unsigned number
// represented by a uint64, whose last 6 digits represent the fractional
// part.
type SigmaMultiplier uint64

func NewSigmaMultiplier(data []byte) (SigmaMultiplier, error) {
	if len(data) != 8 {
		return 0, ErrInvalidSigmaMultiplier.Wrapf("expected 8 bytes, got %d", len(data))
	}

	var s SigmaMultiplier
	err := binary.Read(bytes.NewReader(data), binary.BigEndian, &s)
	if err != nil {
		return s, err
	}
	return s, nil
}

// WholeNumber returns SigmaMultiplier's whole number part.
// For example, the whole number part of SigmaMultiplier(1_000_000) is 1.
func (s SigmaMultiplier) WholeNumber() uint64 {
	return uint64(s) / 1e6
}

// FractionalPart returns SigmaMultiplier's fractional part as a uint64.
// For example, the fractional part of SigmaMultiplier(1_500_000) is 500_000,
// representing the fraction 500_000 / 1_000_000.
func (s SigmaMultiplier) FractionalPart() uint64 {
	return uint64(s) % 1e6
}

// BigRat returns a big.Rat representation of the SigmaMultiplier object.
func (s SigmaMultiplier) BigRat() *big.Rat {
	sigmaInt := new(big.Int).SetUint64(uint64(s))
	return new(big.Rat).SetFrac(sigmaInt, big.NewInt(1e6))
}
